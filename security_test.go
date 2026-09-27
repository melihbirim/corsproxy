package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func testConfig() {
	config = Config{
		MaxRequestSize: 1024 * 1024,
		RequestTimeout: 5 * time.Second,
		MaxRedirects:   5,
		AllowedOrigins: []string{"*"},
	}
}

func TestIsHostAllowed(t *testing.T) {
	testConfig()
	config.BlockedHosts = []string{"blocked.com"}
	cases := map[string]bool{
		"https://api.example.com/x":        true,
		"https://blocked.com/":             false,
		"https://api.blocked.com/":         false,
		"https://notblocked.com.evil.net/": true, // substring match used to block/allow wrongly
		"http://user@blocked.com/":         false,
		"ftp://example.com/":               false,
		"http:///nohost":                   false,
	}
	for u, want := range cases {
		if got := isHostAllowed(u); got != want {
			t.Errorf("isHostAllowed(%q) = %v, want %v", u, got, want)
		}
	}

	config.BlockedHosts = nil
	config.AllowedHosts = []string{"example.com"}
	if !isHostAllowed("https://api.example.com/") || isHostAllowed("https://example.com.evil.net/") {
		t.Error("allowlist must match exact host or subdomain only")
	}
}

func TestIsPrivateIP(t *testing.T) {
	private := []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254",
		"100.64.0.1", "0.0.0.0", "::1", "fd00::1", "fe80::1", "::ffff:127.0.0.1", "::"}
	public := []string{"8.8.8.8", "1.1.1.1", "100.63.255.255", "172.32.0.1", "2606:4700::1111"}
	for _, s := range private {
		if !isPrivateIP(net.ParseIP(s)) {
			t.Errorf("%s should be private", s)
		}
	}
	for _, s := range public {
		if isPrivateIP(net.ParseIP(s)) {
			t.Errorf("%s should be public", s)
		}
	}
}

func TestProxyRefusesLoopbackByDefault(t *testing.T) {
	testConfig()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "s", Value: "1"})
		w.Write([]byte("internal secret"))
	}))
	defer upstream.Close()

	call := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
		return rec
	}

	if rec := call(); rec.Code != http.StatusForbidden {
		t.Fatalf("loopback target: got %d, want 403", rec.Code)
	}

	config.AllowPrivateNetworks = true
	rec := call()
	if rec.Code != http.StatusOK || rec.Body.String() != "internal secret" {
		t.Fatalf("with ALLOW_PRIVATE_NETWORKS: got %d %q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Set-Cookie") != "" {
		t.Error("upstream Set-Cookie must be dropped")
	}
	if rec.Header().Get("Content-Security-Policy") != "sandbox; default-src 'none'" {
		t.Error("proxied responses must be sandboxed")
	}
}

func TestClientIPUsesRightmostForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	if got := getClientIP(r); got != "203.0.113.9" {
		t.Errorf("getClientIP = %q, want the platform-appended 203.0.113.9", got)
	}
}

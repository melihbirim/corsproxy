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
		_, _ = w.Write([]byte("internal secret"))
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
	testConfig()
	config.TrustProxyHeaders = true
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	if got := getClientIP(r); got != "203.0.113.9" {
		t.Errorf("getClientIP = %q, want the platform-appended 203.0.113.9", got)
	}
}

func TestClientIPIgnoresForwardedForByDefault(t *testing.T) {
	testConfig() // TrustProxyHeaders defaults to false
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	r.RemoteAddr = "198.51.100.1:1234"
	if got := getClientIP(r); got != "198.51.100.1" {
		t.Errorf("getClientIP = %q, want RemoteAddr 198.51.100.1 — an untrusted X-Forwarded-For must not be used", got)
	}
}

func TestAPIKeyAuth(t *testing.T) {
	testConfig()
	config.RequireAPIKey = true
	config.APIKeys = []string{"good-key"}

	header := httptest.NewRequest("GET", "/", nil)
	header.Header.Set("X-API-Key", "good-key")
	if !isValidAPIKey(header) {
		t.Error("valid key via header should be accepted")
	}

	query := httptest.NewRequest("GET", "/?apikey=good-key", nil)
	if !isValidAPIKey(query) {
		t.Error("valid key via query param should be accepted")
	}

	wrong := httptest.NewRequest("GET", "/", nil)
	wrong.Header.Set("X-API-Key", "bad-key")
	if isValidAPIKey(wrong) {
		t.Error("wrong key should be rejected")
	}

	missing := httptest.NewRequest("GET", "/", nil)
	if isValidAPIKey(missing) {
		t.Error("missing key should be rejected when RequireAPIKey is on")
	}

	config.RequireAPIKey = false
	if !isValidAPIKey(missing) {
		t.Error("no key should be required when RequireAPIKey is off")
	}
}

func TestProxyRejectsMissingOrWrongAPIKey(t *testing.T) {
	testConfig()
	config.RequireAPIKey = true
	config.APIKeys = []string{"good-key"}
	config.AllowPrivateNetworks = true // httptest servers are on loopback
	defer func() { config.RequireAPIKey = false }()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no key: got %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil)
	req.Header.Set("X-API-Key", "good-key")
	corsProxyHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid key: got %d, want 200", rec.Code)
	}
}

func TestDailyRequestLimit(t *testing.T) {
	testConfig()
	config.DailyRequestLimit = 2
	dailyCount = 0
	dailyResetTime = time.Time{}
	defer func() { config.DailyRequestLimit = 0 }()

	firstOK := checkDailyLimit()
	secondOK := checkDailyLimit()
	if !firstOK || !secondOK {
		t.Fatal("first two requests within the limit should pass")
	}
	if checkDailyLimit() {
		t.Fatal("third request should be refused once the daily ceiling is hit")
	}
}

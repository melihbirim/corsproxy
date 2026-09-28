package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGetEnvHelpers(t *testing.T) {
	t.Run("getEnv", func(t *testing.T) {
		if got := getEnv("UNSET_VAR_XYZ", "default"); got != "default" {
			t.Errorf("getEnv unset = %q, want default", got)
		}
		t.Setenv("SET_VAR_XYZ", "value")
		if got := getEnv("SET_VAR_XYZ", "default"); got != "value" {
			t.Errorf("getEnv set = %q, want value", got)
		}
	})

	t.Run("getEnvInt", func(t *testing.T) {
		if got := getEnvInt("UNSET_INT_XYZ", 7); got != 7 {
			t.Errorf("getEnvInt unset = %d, want 7", got)
		}
		t.Setenv("SET_INT_XYZ", "42")
		if got := getEnvInt("SET_INT_XYZ", 7); got != 42 {
			t.Errorf("getEnvInt set = %d, want 42", got)
		}
		t.Setenv("BAD_INT_XYZ", "not-a-number")
		if got := getEnvInt("BAD_INT_XYZ", 7); got != 7 {
			t.Errorf("getEnvInt invalid = %d, want fallback 7", got)
		}
	})

	t.Run("getEnvInt64", func(t *testing.T) {
		if got := getEnvInt64("UNSET_INT64_XYZ", 100); got != 100 {
			t.Errorf("getEnvInt64 unset = %d, want 100", got)
		}
		t.Setenv("SET_INT64_XYZ", "5000000000")
		if got := getEnvInt64("SET_INT64_XYZ", 100); got != 5000000000 {
			t.Errorf("getEnvInt64 set = %d, want 5000000000", got)
		}
		t.Setenv("BAD_INT64_XYZ", "nope")
		if got := getEnvInt64("BAD_INT64_XYZ", 100); got != 100 {
			t.Errorf("getEnvInt64 invalid = %d, want fallback 100", got)
		}
	})

	t.Run("getEnvDuration", func(t *testing.T) {
		if got := getEnvDuration("UNSET_DUR_XYZ", 5*time.Second); got != 5*time.Second {
			t.Errorf("getEnvDuration unset = %v, want 5s", got)
		}
		t.Setenv("SET_DUR_XYZ", "30s")
		if got := getEnvDuration("SET_DUR_XYZ", 5*time.Second); got != 30*time.Second {
			t.Errorf("getEnvDuration set = %v, want 30s", got)
		}
		t.Setenv("BAD_DUR_XYZ", "nonsense")
		if got := getEnvDuration("BAD_DUR_XYZ", 5*time.Second); got != 5*time.Second {
			t.Errorf("getEnvDuration invalid = %v, want fallback 5s", got)
		}
	})

	t.Run("getEnvBool", func(t *testing.T) {
		if got := getEnvBool("UNSET_BOOL_XYZ", false); got != false {
			t.Errorf("getEnvBool unset = %v, want false", got)
		}
		for _, v := range []string{"true", "1", "yes"} {
			t.Setenv("SET_BOOL_XYZ", v)
			if got := getEnvBool("SET_BOOL_XYZ", false); got != true {
				t.Errorf("getEnvBool(%q) = %v, want true", v, got)
			}
		}
		t.Setenv("SET_BOOL_FALSE_XYZ", "nope")
		if got := getEnvBool("SET_BOOL_FALSE_XYZ", false); got != false {
			t.Errorf("getEnvBool invalid = %v, want false", got)
		}
	})

	t.Run("getEnvList", func(t *testing.T) {
		if got := getEnvList("UNSET_LIST_XYZ", ""); len(got) != 0 {
			t.Errorf("getEnvList unset with empty default = %v, want empty", got)
		}
		t.Setenv("SET_LIST_XYZ", "a, ,b,,c")
		got := getEnvList("SET_LIST_XYZ", "")
		want := []string{"a", "b", "c"}
		if len(got) != len(want) {
			t.Fatalf("getEnvList = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("getEnvList[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "50")
	t.Setenv("ALLOWED_HOSTS", "api.example.com, api.test.com")
	t.Setenv("REQUIRE_API_KEY", "true")
	t.Setenv("API_KEYS", "key1,key2")
	t.Setenv("DAILY_REQUEST_LIMIT", "1000")

	loadConfig()

	if config.Port != "9090" {
		t.Errorf("Port = %q, want 9090", config.Port)
	}
	if config.RateLimitPerMinute != 50 {
		t.Errorf("RateLimitPerMinute = %d, want 50", config.RateLimitPerMinute)
	}
	if len(config.AllowedHosts) != 2 || config.AllowedHosts[0] != "api.example.com" {
		t.Errorf("AllowedHosts = %v", config.AllowedHosts)
	}
	if !config.RequireAPIKey {
		t.Error("RequireAPIKey should be true")
	}
	if len(config.APIKeys) != 2 {
		t.Errorf("APIKeys = %v, want 2 entries", config.APIKeys)
	}
	if config.DailyRequestLimit != 1000 {
		t.Errorf("DailyRequestLimit = %d, want 1000", config.DailyRequestLimit)
	}
}

func TestWarnOnOpenDefaults(t *testing.T) {
	// Fully open: should not panic, exercises every branch. warnOnOpenDefaults
	// only logs, so this is a smoke test that it runs to completion either way.
	testConfig()
	warnOnOpenDefaults()

	config.RateLimitPerMinute = 10
	config.AllowedHosts = []string{"example.com"}
	config.AllowedOrigins = []string{"https://example.com"}
	config.RequireAPIKey = true
	warnOnOpenDefaults()
}

func TestHealthCheckHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	healthCheckHandler(rec, httptest.NewRequest("GET", "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want ok", body["status"])
	}
	if body["timestamp"] == "" {
		t.Error("timestamp field should not be empty")
	}
}

func TestCheckRateLimit(t *testing.T) {
	testConfig()
	config.RateLimitPerMinute = 2
	rateLimiter = make(map[string]*RateLimit)

	ip := "203.0.113.5"
	if !checkRateLimit(ip) || !checkRateLimit(ip) {
		t.Fatal("first two requests within the limit should pass")
	}
	if checkRateLimit(ip) {
		t.Fatal("third request should be refused once the per-minute limit is hit")
	}

	// A different IP has its own independent window.
	if !checkRateLimit("203.0.113.6") {
		t.Error("a different IP should not share the exhausted window")
	}
}

func TestGetClientIPFallsBackToRemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "198.51.100.7:54321"
	if got := getClientIP(r); got != "198.51.100.7" {
		t.Errorf("getClientIP fallback = %q, want 198.51.100.7", got)
	}
}

func TestGetAllowedOrigin(t *testing.T) {
	testConfig()

	t.Run("wildcard allows everything", func(t *testing.T) {
		config.AllowedOrigins = []string{"*"}
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Origin", "https://anything.example.com")
		if got := getAllowedOrigin(r); got != "*" {
			t.Errorf("got %q, want *", got)
		}
	})

	t.Run("no Origin header returns first configured origin", func(t *testing.T) {
		config.AllowedOrigins = []string{"https://a.example.com", "https://b.example.com"}
		r := httptest.NewRequest("GET", "/", nil)
		if got := getAllowedOrigin(r); got != "https://a.example.com" {
			t.Errorf("got %q, want https://a.example.com", got)
		}
	})

	t.Run("matching origin is echoed back", func(t *testing.T) {
		config.AllowedOrigins = []string{"https://a.example.com", "https://b.example.com"}
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Origin", "https://b.example.com")
		if got := getAllowedOrigin(r); got != "https://b.example.com" {
			t.Errorf("got %q, want https://b.example.com", got)
		}
	})

	t.Run("non-matching origin falls back to first configured origin", func(t *testing.T) {
		config.AllowedOrigins = []string{"https://a.example.com"}
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Origin", "https://evil.example.com")
		if got := getAllowedOrigin(r); got != "https://a.example.com" {
			t.Errorf("got %q, want https://a.example.com", got)
		}
	})
}

func TestProxyHandlerRejectsMissingAndInvalidURL(t *testing.T) {
	testConfig()

	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing url: got %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url=ftp://example.com/file", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("non-http(s) scheme: got %d, want 400", rec.Code)
	}
}

func TestProxyHandlerPreflight(t *testing.T) {
	testConfig()
	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("OPTIONS", "/?url="+url.QueryEscape("https://example.com"), nil))
	if rec.Code != http.StatusOK {
		t.Errorf("OPTIONS preflight: got %d, want 200", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Error("preflight response should have an empty body")
	}
}

func TestProxyHandlerRejectsDisallowedHost(t *testing.T) {
	testConfig()
	config.AllowedHosts = []string{"api.example.com"}

	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape("https://not-allowed.example.com/data"), nil))
	if rec.Code != http.StatusForbidden {
		t.Errorf("disallowed host: got %d, want 403", rec.Code)
	}
}

func TestProxyHandlerRateLimited(t *testing.T) {
	testConfig()
	config.RateLimitPerMinute = 1
	config.AllowPrivateNetworks = true // so the request gets past the SSRF check to reach rate limiting
	rateLimiter = make(map[string]*RateLimit)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	call := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil)
		req.RemoteAddr = "203.0.113.9:1234"
		corsProxyHandler(rec, req)
		return rec
	}

	if rec := call(); rec.Code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", rec.Code)
	}
	if rec := call(); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request over the limit: got %d, want 429", rec.Code)
	}
}

func TestProxyHandlerDailyLimitReached(t *testing.T) {
	testConfig()
	config.AllowPrivateNetworks = true
	config.DailyRequestLimit = 0 // disabled: checkDailyLimit always true
	dailyCount = 0
	dailyResetTime = time.Time{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("under disabled limit: got %d, want 200", rec.Code)
	}

	config.DailyRequestLimit = 1
	dailyCount = 1 // already at the ceiling
	dailyResetTime = time.Now().Add(time.Hour)

	rec = httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("at daily ceiling: got %d, want 503", rec.Code)
	}
	config.DailyRequestLimit = 0
}

func TestProxyHandlerSetsCredentialsHeaderForSpecificOrigin(t *testing.T) {
	testConfig()
	config.AllowedOrigins = []string{"https://app.example.com"}
	config.AllowPrivateNetworks = true

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil)
	req.Header.Set("Origin", "https://app.example.com")
	corsProxyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected Access-Control-Allow-Credentials: true for a non-wildcard origin")
	}
}

func TestProxyHandlerRejectsOversizedUpstreamResponse(t *testing.T) {
	testConfig()
	config.AllowPrivateNetworks = true
	config.MaxRequestSize = 10 // bytes: small enough that any real response trips it

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this response body is well over ten bytes long"))
	}))
	defer upstream.Close()

	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("oversized upstream response: got %d, want 502", rec.Code)
	}
}

func TestProxyHandlerRedirectHandling(t *testing.T) {
	testConfig()
	config.AllowPrivateNetworks = true

	t.Run("too many redirects", func(t *testing.T) {
		config.MaxRedirects = 0
		var upstream *httptest.Server
		upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, upstream.URL+"/next", http.StatusFound)
		}))
		defer upstream.Close()

		rec := httptest.NewRecorder()
		corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
		if rec.Code != http.StatusBadGateway {
			t.Errorf("too many redirects: got %d, want 502", rec.Code)
		}
	})

	t.Run("redirect to a disallowed host is refused", func(t *testing.T) {
		config.MaxRedirects = 5
		config.AllowedHosts = []string{"allowed.invalid"}
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://not-allowed.invalid/", http.StatusFound)
		}))
		defer upstream.Close()
		upstreamHost := strings.Split(strings.Split(upstream.URL, "//")[1], ":")[0]
		config.AllowedHosts = append(config.AllowedHosts, upstreamHost)

		rec := httptest.NewRecorder()
		corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
		if rec.Code != http.StatusBadGateway {
			t.Errorf("redirect to disallowed host: got %d, want 502", rec.Code)
		}
		config.AllowedHosts = nil
	})
}

func TestProxyHandlerVerboseLogging(t *testing.T) {
	testConfig()
	config.AllowPrivateNetworks = true
	config.EnableVerboseLog = true
	config.AllowedHosts = []string{"only-this-host.invalid"}
	defer func() { config.EnableVerboseLog = false }()

	// Blocked-host path with verbose logging on.
	rec := httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape("https://not-this-host.invalid/"), nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("blocked host: got %d, want 403", rec.Code)
	}

	// Successful round trip with verbose logging on.
	config.AllowedHosts = nil
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	rec = httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape(upstream.URL), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("successful round trip: got %d, want 200", rec.Code)
	}

	// Fetch failure (connection refused) with verbose logging on.
	rec = httptest.NewRecorder()
	corsProxyHandler(rec, httptest.NewRequest("GET", "/?url="+url.QueryEscape("http://127.0.0.1:1/"), nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unreachable upstream: got %d, want 502", rec.Code)
	}
}

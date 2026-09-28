package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Configuration with defaults
type Config struct {
	Port               string
	MaxRequestSize     int64
	RequestTimeout     time.Duration
	MaxRedirects       int
	AllowedOrigins     []string
	BlockedHosts       []string
	AllowedHosts       []string
	EnableVerboseLog   bool
	RateLimitPerMinute int
	// AllowPrivateNetworks lets the proxy reach loopback/private/link-local
	// addresses. Off by default: an open proxy that can reach them is an
	// SSRF hole (cloud metadata, internal admin panels).
	AllowPrivateNetworks bool
	// RequireAPIKey gates every proxied request behind one of APIKeys.
	// Off by default so the zero-config quick start keeps working.
	RequireAPIKey bool
	APIKeys       []string
	// DailyRequestLimit is a global (not per-IP) kill switch on total
	// proxied requests per rolling 24h window. 0 disables it.
	DailyRequestLimit int
}

var (
	config      Config
	rateLimiter = make(map[string]*RateLimit)
	rateMutex   sync.RWMutex

	dailyCount     int64
	dailyResetTime time.Time
	dailyMutex     sync.Mutex
)

type RateLimit struct {
	count     int
	resetTime time.Time
}

func main() {
	loadConfig()

	http.HandleFunc("/", corsProxyHandler)
	http.HandleFunc("/health", healthCheckHandler)

	log.Printf("🚀 CORS Proxy server starting on port %s", config.Port)
	log.Printf("📝 Usage: http://localhost:%s/?url=https://example.com", config.Port)
	log.Printf("⚙️  Max request size: %d MB", config.MaxRequestSize/(1024*1024))
	log.Printf("⏱️  Request timeout: %v", config.RequestTimeout)
	if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
		log.Printf("🌐 CORS: All origins allowed (*)")
	} else {
		log.Printf("🌐 CORS: Specific origins allowed: %v", config.AllowedOrigins)
	}
	if config.RateLimitPerMinute > 0 {
		log.Printf("🚦 Rate limit: %d requests/minute per IP", config.RateLimitPerMinute)
	}
	if len(config.AllowedHosts) > 0 {
		log.Printf("✅ Allowed hosts: %v", config.AllowedHosts)
	}
	if len(config.BlockedHosts) > 0 {
		log.Printf("🚫 Blocked hosts: %v", config.BlockedHosts)
	}
	if config.RequireAPIKey {
		log.Printf("🔑 API key required (%d key(s) configured)", len(config.APIKeys))
	}
	if config.DailyRequestLimit > 0 {
		log.Printf("🛑 Daily request limit: %d", config.DailyRequestLimit)
	}
	warnOnOpenDefaults()

	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func loadConfig() {
	config = Config{
		Port:                 getEnv("PORT", "8080"),
		MaxRequestSize:       getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB default
		RequestTimeout:       getEnvDuration("REQUEST_TIMEOUT", 30*time.Second),
		MaxRedirects:         getEnvInt("MAX_REDIRECTS", 10),
		AllowedOrigins:       getEnvList("ALLOWED_ORIGINS", "*"),
		BlockedHosts:         getEnvList("BLOCKED_HOSTS", ""),
		AllowedHosts:         getEnvList("ALLOWED_HOSTS", ""),
		EnableVerboseLog:     getEnvBool("VERBOSE_LOGGING", false),
		AllowPrivateNetworks: getEnvBool("ALLOW_PRIVATE_NETWORKS", false),
		RateLimitPerMinute:   getEnvInt("RATE_LIMIT_PER_MINUTE", 0), // 0 = disabled
		RequireAPIKey:        getEnvBool("REQUIRE_API_KEY", false),
		APIKeys:              getEnvList("API_KEYS", ""),
		DailyRequestLimit:    getEnvInt("DAILY_REQUEST_LIMIT", 0), // 0 = disabled
	}
}

// warnOnOpenDefaults logs a single startup warning listing which safety
// nets are off, so a wide-open deployment isn't silent about it.
func warnOnOpenDefaults() {
	var open []string
	if config.RateLimitPerMinute == 0 {
		open = append(open, "no per-IP rate limit (RATE_LIMIT_PER_MINUTE=0)")
	}
	if len(config.AllowedHosts) == 0 {
		open = append(open, "no destination allowlist (ALLOWED_HOSTS empty, any public host can be proxied)")
	}
	if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
		open = append(open, "any origin can call it (ALLOWED_ORIGINS=*)")
	}
	if !config.RequireAPIKey {
		open = append(open, "no API key required (REQUIRE_API_KEY=false)")
	}
	if len(open) == 0 {
		return
	}
	log.Printf("⚠️  Running with open defaults: %s. Fine for local dev; see README's Production checklist before exposing this publicly.", strings.Join(open, "; "))
}

// isValidAPIKey reports whether the request is authorized. Always true
// when RequireAPIKey is off (the zero-config default).
func isValidAPIKey(r *http.Request) bool {
	if !config.RequireAPIKey {
		return true
	}
	provided := r.Header.Get("X-API-Key")
	if provided == "" {
		provided = r.URL.Query().Get("apikey")
	}
	if provided == "" {
		return false
	}
	for _, key := range config.APIKeys {
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) == 1 {
			return true
		}
	}
	return false
}

// checkDailyLimit enforces a global (all clients combined) request ceiling
// over a rolling 24h window, as a cost/abuse kill switch independent of the
// per-IP rate limiter. Always true when DailyRequestLimit is 0.
func checkDailyLimit() bool {
	if config.DailyRequestLimit <= 0 {
		return true
	}
	dailyMutex.Lock()
	defer dailyMutex.Unlock()
	now := time.Now()
	if dailyResetTime.IsZero() || now.After(dailyResetTime) {
		dailyCount = 0
		dailyResetTime = now.Add(24 * time.Hour)
	}
	if dailyCount >= int64(config.DailyRequestLimit) {
		return false
	}
	dailyCount++
	return true
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func getEnvList(key, defaultVal string) []string {
	val := getEnv(key, defaultVal)
	if val == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339)); err != nil {
		log.Printf("Error writing health check response: %v", err)
	}
}

func corsProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Determine which origin to allow based on request Origin header
	allowedOrigin := getAllowedOrigin(r)

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Add credentials header if not wildcard
	if allowedOrigin != "*" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Global daily request ceiling (kill switch), checked before anything
	// else costs work.
	if !checkDailyLimit() {
		http.Error(w, `{"error":"Daily request limit reached. Try again tomorrow."}`, http.StatusServiceUnavailable)
		return
	}

	// Optional API key authentication
	if !isValidAPIKey(r) {
		http.Error(w, `{"error":"Missing or invalid API key"}`, http.StatusUnauthorized)
		return
	}

	// Rate limiting
	if config.RateLimitPerMinute > 0 {
		clientIP := getClientIP(r)
		if !checkRateLimit(clientIP) {
			http.Error(w, `{"error":"Rate limit exceeded. Please try again later."}`, http.StatusTooManyRequests)
			return
		}
	}

	// Get target URL from query parameter
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error":"Missing 'url' parameter. Usage: /?url=https://example.com"}`, http.StatusBadRequest)
		return
	}

	// Validate URL
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		http.Error(w, `{"error":"URL must start with http:// or https://"}`, http.StatusBadRequest)
		return
	}

	// Check allowed/blocked hosts
	if !isHostAllowed(targetURL) {
		http.Error(w, `{"error":"This host is not allowed"}`, http.StatusForbidden)
		if config.EnableVerboseLog {
			log.Printf("🚫 Blocked request to: %s", targetURL)
		}
		return
	}

	// Cap the request body like the response.
	r.Body = http.MaxBytesReader(w, r.Body, config.MaxRequestSize)

	// Create new request
	proxyReq, err := http.NewRequestWithContext(context.Background(), r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"Invalid URL: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Copy headers from original request (except Host)
	for key, values := range r.Header {
		if key != "Host" {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Make the request
	client := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           safeDialer().DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: config.RequestTimeout,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			if !isHostAllowed(req.URL.String()) {
				return fmt.Errorf("redirect to a host that is not allowed")
			}
			return nil
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		if config.EnableVerboseLog {
			log.Printf("Error fetching URL %s: %v", targetURL, err)
		}
		if errors.Is(err, errPrivateAddress) {
			http.Error(w, `{"error":"This host is not allowed"}`, http.StatusForbidden)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"Failed to fetch URL: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// Refuse oversized responses up front instead of truncating them silently.
	if resp.ContentLength > config.MaxRequestSize {
		http.Error(w, `{"error":"Upstream response is too large"}`, http.StatusBadGateway)
		return
	}

	// Copy response headers, except cookies: with credentials allowed, an
	// upstream could otherwise set cookies on the proxy's own domain.
	for key, values := range resp.Header {
		if strings.EqualFold(key, "Set-Cookie") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	// Proxied content must never run as a page on the proxy's origin
	// (phishing/XSS). fetch()/XHR callers are unaffected.
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Override CORS headers
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	if allowedOrigin != "*" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body with size limit
	limitedReader := io.LimitReader(resp.Body, config.MaxRequestSize)
	written, err := io.Copy(w, limitedReader)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
		return
	}

	if config.EnableVerboseLog {
		log.Printf("%s %s -> %s (%d bytes, %d status)", r.Method, targetURL, resp.Status, written, resp.StatusCode)
	}
}

func getClientIP(r *http.Request) string {
	// Behind a platform proxy (Railway, Render, Fly), the rightmost
	// X-Forwarded-For entry is the one the platform appended; entries to its
	// left come from the client and can be forged to dodge rate limits.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[len(ips)-1])
	}
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func checkRateLimit(clientIP string) bool {
	rateMutex.Lock()
	defer rateMutex.Unlock()

	now := time.Now()
	// Drop expired windows once the map grows, so spoofed or rotating IPs
	// can't grow it without bound.
	if len(rateLimiter) > 10000 {
		for ip, entry := range rateLimiter {
			if now.After(entry.resetTime) {
				delete(rateLimiter, ip)
			}
		}
	}
	rl, exists := rateLimiter[clientIP]

	if !exists || now.After(rl.resetTime) {
		// Create new rate limit window
		rateLimiter[clientIP] = &RateLimit{
			count:     1,
			resetTime: now.Add(time.Minute),
		}
		return true
	}

	if rl.count >= config.RateLimitPerMinute {
		return false
	}

	rl.count++
	return true
}

// hostMatches reports whether host is rule or a subdomain of it
// ("example.com" matches "api.example.com", not "example.com.evil.net").
func hostMatches(host, rule string) bool {
	rule = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(rule), "."))
	return rule != "" && (host == rule || strings.HasSuffix(host, "."+rule))
}

func isHostAllowed(targetURL string) bool {
	u, err := url.Parse(targetURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return false
	}
	host := strings.ToLower(u.Hostname()) // strips userinfo, port, IPv6 brackets

	for _, blocked := range config.BlockedHosts {
		if hostMatches(host, blocked) {
			return false
		}
	}
	if len(config.AllowedHosts) == 0 {
		return true
	}
	for _, allowed := range config.AllowedHosts {
		if hostMatches(host, allowed) {
			return true
		}
	}
	return false
}

// isPrivateIP covers loopback, RFC 1918, link-local (incl. cloud metadata
// 169.254.169.254), CGNAT, unspecified, multicast, and IPv6 ULA/link-local.
// IPv4-mapped IPv6 is unwrapped first.
func isPrivateIP(ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		ip = v4
		if v4[0] == 0 || (v4[0] == 100 && v4[1]&0xc0 == 64) { // 0.0.0.0/8, 100.64.0.0/10
			return true
		}
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast()
}

var errPrivateAddress = errors.New("destination is a private or internal address")

// safeDialer checks the address actually being connected to, after DNS
// resolution, on every connection (including after redirects). That also
// defeats DNS rebinding, which a pre-request hostname check can't.
func safeDialer() *net.Dialer {
	d := &net.Dialer{Timeout: 10 * time.Second}
	if config.AllowPrivateNetworks {
		return d
	}
	d.Control = func(network, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		if ip := net.ParseIP(host); ip == nil || isPrivateIP(ip) {
			return errPrivateAddress
		}
		return nil
	}
	return d
}

func getAllowedOrigin(r *http.Request) string {
	// If wildcard, allow all
	if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
		return "*"
	}

	// Get the Origin header from the request
	requestOrigin := r.Header.Get("Origin")
	if requestOrigin == "" {
		// No Origin header, return first allowed origin or *
		if len(config.AllowedOrigins) > 0 {
			return config.AllowedOrigins[0]
		}
		return "*"
	}

	// Check if the request origin is in our allowed list
	for _, allowed := range config.AllowedOrigins {
		if allowed == "*" || requestOrigin == allowed {
			return requestOrigin
		}
	}

	// If not found in allowed list, return the first allowed origin
	// This will cause CORS to fail on the browser side
	if len(config.AllowedOrigins) > 0 {
		return config.AllowedOrigins[0]
	}
	return "*"
}

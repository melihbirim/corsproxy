package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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
}

var (
	config      Config
	rateLimiter = make(map[string]*RateLimit)
	rateMutex   sync.RWMutex
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

	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func loadConfig() {
	config = Config{
		Port:               getEnv("PORT", "8080"),
		MaxRequestSize:     getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB default
		RequestTimeout:     getEnvDuration("REQUEST_TIMEOUT", 30*time.Second),
		MaxRedirects:       getEnvInt("MAX_REDIRECTS", 10),
		AllowedOrigins:     getEnvList("ALLOWED_ORIGINS", "*"),
		BlockedHosts:       getEnvList("BLOCKED_HOSTS", ""),
		AllowedHosts:       getEnvList("ALLOWED_HOSTS", ""),
		EnableVerboseLog:   getEnvBool("VERBOSE_LOGGING", false),
		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 0), // 0 = disabled
	}
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

	// Create new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		if config.EnableVerboseLog {
			log.Printf("Error fetching URL %s: %v", targetURL, err)
		}
		http.Error(w, fmt.Sprintf(`{"error":"Failed to fetch URL: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

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
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
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

func isHostAllowed(targetURL string) bool {
	// Extract host from URL
	host := targetURL
	if idx := strings.Index(targetURL, "://"); idx != -1 {
		host = targetURL[idx+3:]
	}
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Check blocked hosts first
	for _, blocked := range config.BlockedHosts {
		if strings.Contains(host, blocked) {
			return false
		}
	}

	// If allowed hosts is empty, allow all (except blocked)
	if len(config.AllowedHosts) == 0 {
		return true
	}

	// Check if host is in allowed list
	for _, allowed := range config.AllowedHosts {
		if strings.Contains(host, allowed) {
			return true
		}
	}

	return false
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

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultPort    = "8080"
	maxRequestSize = 10 * 1024 * 1024 // 10MB
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", corsProxyHandler)
	http.HandleFunc("/health", healthCheckHandler)

	log.Printf("🚀 CORS Proxy server starting on port %s", port)
	log.Printf("📝 Usage: http://localhost:%s/?url=https://example.com", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

func corsProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
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
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error fetching URL %s: %v", targetURL, err)
		http.Error(w, fmt.Sprintf(`{"error":"Failed to fetch URL: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Override CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body with size limit
	limitedReader := io.LimitReader(resp.Body, maxRequestSize)
	written, err := io.Copy(w, limitedReader)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
		return
	}

	log.Printf("%s %s -> %s (%d bytes, %d status)", r.Method, targetURL, resp.Status, written, resp.StatusCode)
}

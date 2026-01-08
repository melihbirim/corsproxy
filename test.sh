#!/bin/bash

# CORS Proxy Test Script
# Tests the proxy with various endpoints and scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROXY_URL="${PROXY_URL:-http://localhost:8080}"
PROXY_PORT="${PROXY_PORT:-8080}"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SERVER_PID=""

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up...${NC}"
    
    if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    
    # Kill any remaining processes on the port
    if lsof -ti:$PROXY_PORT > /dev/null 2>&1; then
        echo "Cleaning up port $PROXY_PORT..."
        lsof -ti:$PROXY_PORT | xargs kill -9 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}TEST $TOTAL_TESTS:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓ PASSED:${NC} $1\n"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

print_failure() {
    echo -e "${RED}✗ FAILED:${NC} $1\n"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_test "$1"
}

# Check if server is running
check_server() {
    if ! curl -s -f "${PROXY_URL}/health" > /dev/null 2>&1; then
        return 1
    fi
    return 0
}

# Stop any existing server on the port
stop_existing_server() {
    if lsof -ti:$PROXY_PORT > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Found existing process on port $PROXY_PORT${NC}"
        echo "Stopping existing server..."
        lsof -ti:$PROXY_PORT | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
}

# Start the server
start_server() {
    echo -e "${BLUE}🚀 Starting CORS Proxy server...${NC}"
    
    # Check if binary exists
    if [ -f "./bin/cors-proxy" ]; then
        ./bin/cors-proxy > /tmp/cors-proxy-test.log 2>&1 &
        SERVER_PID=$!
    elif [ -f "./cors-proxy" ]; then
        ./cors-proxy > /tmp/cors-proxy-test.log 2>&1 &
        SERVER_PID=$!
    else
        # Build and run
        echo "Building server..."
        go build -o /tmp/cors-proxy-test main.go
        /tmp/cors-proxy-test > /tmp/cors-proxy-test.log 2>&1 &
        SERVER_PID=$!
    fi
    
    echo "Server started with PID: $SERVER_PID"
    
    # Wait for server to be ready
    echo "Waiting for server to be ready..."
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        if check_server; then
            echo -e "${GREEN}✓ Server is ready${NC}\n"
            return 0
        fi
        attempts=$((attempts + 1))
        sleep 1
    done
    
    echo -e "${RED}✗ Server failed to start${NC}"
    echo "Server log:"
    cat /tmp/cors-proxy-test.log
    exit 1
}

# Test 1: Health Check
test_health() {
    run_test "Health check endpoint"
    
    response=$(curl -s "${PROXY_URL}/health")
    
    if echo "$response" | grep -q '"status":"ok"'; then
        print_success "Health check returned OK status"
    else
        print_failure "Health check failed. Response: $response"
    fi
}

# Test 2: Missing URL parameter
test_missing_url() {
    run_test "Request without URL parameter (should fail)"
    
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY_URL}/")
    
    if [ "$http_code" = "400" ]; then
        print_success "Correctly returned 400 for missing URL"
    else
        print_failure "Expected 400, got $http_code"
    fi
}

# Test 3: Invalid URL (no protocol)
test_invalid_url() {
    run_test "Request with invalid URL (no http/https)"
    
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY_URL}/?url=example.com")
    
    if [ "$http_code" = "400" ]; then
        print_success "Correctly rejected URL without protocol"
    else
        print_failure "Expected 400, got $http_code"
    fi
}

# Test 4: GitHub API (JSON)
test_github_api() {
    run_test "Fetch GitHub API (JSON response)"
    
    response=$(curl -s "${PROXY_URL}/?url=https://api.github.com/users/melihbirim")
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY_URL}/?url=https://api.github.com/users/melihbirim")
    
    if [ "$http_code" = "200" ] && echo "$response" | grep -q '"login"'; then
        print_success "Successfully fetched GitHub user data"
        echo "Sample response: $(echo $response | head -c 100)..."
    else
        print_failure "Failed to fetch GitHub data. HTTP: $http_code"
    fi
}

# Test 5: Gutenberg Moby Dick (Large text file)
test_moby_dick() {
    run_test "Fetch Moby Dick from Gutenberg (large text file)"
    
    response=$(curl -s "${PROXY_URL}/?url=https://www.gutenberg.org/cache/epub/2701/pg2701.txt")
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY_URL}/?url=https://www.gutenberg.org/cache/epub/2701/pg2701.txt")
    
    if [ "$http_code" = "200" ] && echo "$response" | grep -q "Moby Dick"; then
        word_count=$(echo "$response" | wc -w)
        print_success "Successfully fetched Moby Dick ($word_count words)"
    else
        print_failure "Failed to fetch Moby Dick. HTTP: $http_code"
    fi
}

# Test 6: CORS Headers
test_cors_headers() {
    run_test "Check CORS headers"
    
    headers=$(curl -s -I "${PROXY_URL}/?url=https://api.github.com")
    
    if echo "$headers" | grep -qi "Access-Control-Allow-Origin"; then
        print_success "CORS headers present"
    else
        print_failure "CORS headers missing"
    fi
}

# Test 7: OPTIONS (Preflight) Request
test_preflight() {
    run_test "OPTIONS preflight request"
    
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -X OPTIONS "${PROXY_URL}/?url=https://example.com")
    
    if [ "$http_code" = "200" ]; then
        print_success "Preflight request handled correctly"
    else
        print_failure "Preflight failed. Expected 200, got $http_code"
    fi
}

# Test 8: POST Request with JSON
test_post_json() {
    run_test "POST request with JSON body"
    
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"title":"Test"}' \
        "${PROXY_URL}/?url=https://httpbin.org/post")
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d '{"title":"Test"}' \
        "${PROXY_URL}/?url=https://httpbin.org/post")
    
    if [ "$http_code" = "200" ] && echo "$response" | grep -q '"title"'; then
        print_success "POST request successful"
    else
        print_failure "POST request failed. HTTP: $http_code"
    fi
}

# Test 9: Custom Headers
test_custom_headers() {
    run_test "Request with custom headers"
    
    response=$(curl -s \
        -H "X-Custom-Header: TestValue" \
        "${PROXY_URL}/?url=https://httpbin.org/headers")
    
    if echo "$response" | grep -q "X-Custom-Header"; then
        print_success "Custom headers forwarded correctly"
    else
        print_failure "Custom headers not forwarded"
    fi
}

# Test 10: Response Time
test_response_time() {
    run_test "Response time check"
    
    start_time=$(date +%s%N)
    curl -s "${PROXY_URL}/?url=https://api.github.com" > /dev/null
    end_time=$(date +%s%N)
    
    duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
    
    if [ "$duration" -lt 5000 ]; then
        print_success "Response time: ${duration}ms (under 5s)"
    else
        print_failure "Response time too slow: ${duration}ms"
    fi
}

# Main execution
main() {
    print_header "🧪 CORS Proxy Test Suite"
    
    echo "Testing proxy at: $PROXY_URL"
    echo ""
    
    # Stop any existing server and start fresh
    stop_existing_server
    start_server
    
    print_header "Running Tests"
    
    # Run all tests
    test_health
    test_missing_url
    test_invalid_url
    test_github_api
    test_moby_dick
    test_cors_headers
    test_preflight
    test_post_json
    test_custom_headers
    test_response_time
    
    # Summary
    print_header "Test Summary"
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
    else
        echo -e "Failed:       $FAILED_TESTS"
    fi
    echo ""
    
    if [ "$FAILED_TESTS" -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}\n"
        exit 0
    else
        echo -e "${RED}❌ Some tests failed${NC}\n"
        exit 1
    fi
}

# Run main function
main "$@"

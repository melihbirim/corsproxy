# Test Coverage Analysis - CORS Proxy

**Date:** May 18, 2026  
**Project:** CORS Proxy - Open Source Edition  
**Status:** ⚠️ Needs Improvement

## Executive Summary

The CORS Proxy project currently has **only integration tests** (shell script) with no unit tests. While the integration tests cover happy paths, there are significant gaps in test coverage that leave many critical functions and edge cases untested.

### Current State
- **Integration Tests:** 10 tests (test.sh)
- **Unit Tests:** 0 tests
- **Code Coverage:** Unknown (likely <20%)
- **Test Framework:** Bash script (no Go testing)

---

## Coverage Analysis by Function

### ✅ Functions with Integration Test Coverage

| Function | Coverage | Depth |
|----------|----------|-------|
| `corsProxyHandler` | ✅ Partial | Happy path only |
| `healthCheckHandler` | ✅ Full | All paths covered |
| `getAllowedOrigin` | ✅ Partial | CORS header checks only |
| `isHostAllowed` | ❌ None | Not directly tested |
| `checkRateLimit` | ❌ None | Not directly tested |
| `getClientIP` | ❌ None | Not directly tested |
| Configuration Functions | ❌ None | Not tested at all |

---

## Critical Gaps in Test Coverage

### 1. **Configuration Loading & Parsing** (HIGH PRIORITY)
**Functions:** `loadConfig`, `getEnv*` functions
**Issues:**
- No tests for invalid environment variables
- No tests for type conversion edge cases (negative numbers, oversized integers)
- No tests for malformed duration strings
- No tests for empty/null values
- No tests for list parsing with various delimiters

**Scenarios Missing:**
```go
- getEnvInt with non-numeric string
- getEnvInt64 with overflow values
- getEnvDuration with invalid duration format
- getEnvBool with values other than "true", "1", "yes"
- getEnvList with empty strings, special characters
- loadConfig with conflicting or invalid settings
```

### 2. **Rate Limiting Logic** (HIGH PRIORITY)
**Functions:** `checkRateLimit`
**Issues:**
- No tests for concurrent access (thread safety)
- No tests for rate limit window expiration
- No tests for cleanup of expired entries
- No tests for edge case: exactly at the limit
- No tests for memory leaks (rateLimiter map growth)

**Scenarios Missing:**
```go
- Multiple IPs hitting rate limit simultaneously
- Rate window boundary conditions (at 59s, 60s, 61s)
- Cleanup of old rate limit entries
- Behavior when RateLimitPerMinute = 0 (disabled)
- Large number of unique IPs over time
```

### 3. **Host Validation** (HIGH PRIORITY)
**Functions:** `isHostAllowed`
**Issues:**
- No tests for port stripping edge cases
- No tests for subdomain matching
- No tests with IPv6 addresses
- No tests for URLs with authentication (user:pass@host)
- No tests for international domain names (IDN)

**Scenarios Missing:**
```go
- Domain with multiple ports (malformed)
- Subdomain matching logic
- IPv6 URLs: https://[::1]:8080/path
- URLs with userinfo: https://user:pass@example.com
- Case sensitivity of domains
- Wildcard domain patterns
- Special characters in domains
```

### 4. **Error Handling** (MEDIUM PRIORITY)
**Functions:** All handler functions
**Issues:**
- No tests for network timeouts
- No tests for response body read errors
- No tests for request creation failures
- No tests for header setting failures
- No tests for malformed URLs

**Scenarios Missing:**
```go
- Timeout during client request
- Incomplete response body (connection drop)
- Invalid redirect chains
- Too many redirects (exceeds MaxRedirects)
- Missing response body closer
- Response header overwrites
```

### 5. **CORS Header Logic** (MEDIUM PRIORITY)
**Functions:** `getAllowedOrigin`, `corsProxyHandler` (header section)
**Issues:**
- No tests for wildcard origin configuration
- No tests for origin mismatch scenarios
- No tests for credentials header with wildcard
- No tests for multiple allowed origins
- No tests for empty origin header

**Scenarios Missing:**
```go
- AllowedOrigins = ["https://example.com", "https://other.com"]
- Request Origin not in AllowedOrigins
- Origin header missing from request
- Credentials flag with/without wildcard
- Special origins (null, file://)
- Origin header with port numbers
```

### 6. **HTTP Header Handling** (MEDIUM PRIORITY)
**Functions:** `corsProxyHandler` (header copy sections)
**Issues:**
- No tests for case-sensitive header handling
- No tests for multi-value headers
- No tests for header injection attacks
- No tests for restricted headers (Host, Content-Length, etc.)
- No tests for header overflow

**Scenarios Missing:**
```go
- Headers with multiple values (Set-Cookie)
- Large header values
- Special characters in headers
- HTTP/2 pseudo-headers (if applicable)
- Forbidden header names
- Custom header forwarding
```

### 7. **Proxy Request Creation** (MEDIUM PRIORITY)
**Functions:** Request creation in `corsProxyHandler`
**Issues:**
- No tests for request body size limits
- No tests for streamed request bodies
- No tests for malformed request bodies
- No tests for request body reset

**Scenarios Missing:**
```go
- POST with large body (near MaxRequestSize)
- Request body that exceeds MaxRequestSize
- Non-seekable request body
- Chunked transfer encoding
- Request body encoding issues
```

### 8. **Response Size Limiting** (MEDIUM PRIORITY)
**Functions:** Response body copy with `io.LimitReader`
**Issues:**
- No tests for partial response due to size limit
- No tests for corrupted response when truncated
- No tests for client handling of truncated response
- No error handling when limit reached

**Scenarios Missing:**
```go
- Response exactly at MaxRequestSize
- Response slightly over MaxRequestSize
- Binary data truncation
- Streaming responses
- Large file downloads
```

### 9. **Client IP Detection** (LOW PRIORITY)
**Functions:** `getClientIP`
**Issues:**
- No tests for IPv6 addresses
- No tests for malformed X-Forwarded-For headers
- No tests for port parsing with IPv6
- No tests for empty header values

**Scenarios Missing:**
```go
- X-Forwarded-For: "2001:db8::1, 192.0.2.1"
- X-Forwarded-For with ports
- Malformed RemoteAddr
- X-Real-IP vs X-Forwarded-For precedence
- Multiple values in X-Forwarded-For
```

### 10. **Concurrent/Safety Issues** (LOW PRIORITY)
**General Issues:**
- No tests for race conditions (rateLimiter map access)
- No tests for goroutine leaks
- No tests for proper mutex usage
- No tests for request/response handling under load

---

## Integration Test Coverage Summary

### Current Tests (test.sh)

| # | Test | Coverage Type | Gaps |
|---|------|---------------|------|
| 1 | Health check | ✅ Basic | No error scenarios |
| 2 | Missing URL parameter | ✅ Validation | No edge cases |
| 3 | Invalid URL | ✅ Validation | No protocol variations |
| 4 | GitHub API fetch | ✅ Happy path | No error scenarios |
| 5 | Large file (Moby Dick) | ✅ Happy path | No failure scenarios |
| 6 | CORS headers | ✅ Basic | No origin mismatches |
| 7 | OPTIONS preflight | ✅ Basic | No custom headers |
| 8 | POST request | ✅ Happy path | No error bodies, large payloads |
| 9 | Custom headers | ✅ Happy path | No restricted headers |
| 10 | Response time | ✅ Basic | No timeout scenarios |

**Missing Integration Test Scenarios:**
- ❌ Rate limiting enforcement
- ❌ Blocked/allowed hosts
- ❌ Request timeout behavior
- ❌ Redirect limit enforcement
- ❌ Large request body handling
- ❌ Connection errors
- ❌ Malformed responses
- ❌ Configuration variations

---

## Recommended Improvements by Priority

### 🔴 HIGH PRIORITY (Critical)

#### 1. **Add Unit Tests for Configuration** (Effort: Medium)
- Test all `getEnv*` functions with valid/invalid inputs
- Test type conversions and edge cases
- Test list parsing with various formats
- **Files to create:** `main_config_test.go`
- **Expected coverage:** 15 test cases

#### 2. **Add Unit Tests for Rate Limiting** (Effort: High)
- Test `checkRateLimit` function with various scenarios
- Test concurrent access and thread safety
- Test rate window expiration
- Test with disabled rate limiting (0 value)
- **Files to create:** `main_ratelimit_test.go`
- **Expected coverage:** 20+ test cases
- **Note:** May require refactoring for better testability (extracting time dependency)

#### 3. **Add Unit Tests for Host Validation** (Effort: Medium)
- Test `isHostAllowed` with various URL formats
- Test domain extraction edge cases
- Test blocked/allowed host logic
- **Files to create:** `main_validation_test.go`
- **Expected coverage:** 15+ test cases

### 🟡 MEDIUM PRIORITY (Important)

#### 4. **Add Unit Tests for CORS Logic** (Effort: Medium)
- Test `getAllowedOrigin` with various configurations
- Test origin validation and header setting
- **Files to create:** `main_cors_test.go`
- **Expected coverage:** 12+ test cases

#### 5. **Add Error Scenario Integration Tests** (Effort: High)
- Extend test.sh to cover timeout scenarios
- Test redirect limits
- Test rate limiting enforcement
- Test blocked hosts
- **Expected new tests:** 8+ tests

#### 6. **Add Test Coverage Tool** (Effort: Low)
- Configure `go test -cover` in CI
- Set coverage threshold (target: 60%+)
- Generate coverage reports
- **Files to create:** GitHub Actions workflow, .coverprofile

### 🟢 LOW PRIORITY (Nice to Have)

#### 7. **Add Benchmark Tests** (Effort: Low)
- Performance benchmarks for hot paths
- Rate limiter performance
- Request processing speed
- **Files to create:** `*_bench_test.go`

#### 8. **Add Fuzzing Tests** (Effort: Medium)
- Fuzz URL parsing
- Fuzz header values
- Fuzz configuration strings
- **Files to create:** `*_fuzz_test.go`

---

## Implementation Roadmap

### Phase 1: Foundational Unit Tests (Week 1)
1. Create `main_config_test.go` - Configuration parsing
2. Create `main_validation_test.go` - Host and URL validation
3. Create `main_cors_test.go` - CORS header logic

### Phase 2: Advanced Unit Tests (Week 2)
1. Create `main_ratelimit_test.go` - Rate limiting (may require refactoring)
2. Extend integration tests for error scenarios

### Phase 3: CI/CD Integration (Week 3)
1. Set up GitHub Actions for test execution
2. Configure coverage reporting
3. Add coverage threshold checks

### Phase 4: Performance & Fuzzing (Week 4)
1. Add benchmark tests
2. Add fuzzing tests

---

## Code Quality Improvements for Better Testability

### 1. **Extract Time Dependency** (For Rate Limiting)
Current code uses `time.Now()` directly, which makes testing difficult.
```go
// Current
rl.resetTime = now.Add(time.Minute)

// Suggested refactoring
type TimeProvider interface {
    Now() time.Time
}
```

### 2. **Extract HTTP Client** (For Handler)
Current code creates HTTP client inline, making mocking difficult.
```go
// Suggested refactoring
type Proxy struct {
    client *http.Client
}
```

### 3. **Separate Configuration Validation**
Move validation logic out of `loadConfig` into testable functions.

---

## Summary Table

| Category | Status | Test Count | Priority |
|----------|--------|-----------|----------|
| **Configuration** | ❌ Untested | 0 | 🔴 HIGH |
| **Rate Limiting** | ❌ Untested | 0 | 🔴 HIGH |
| **Host Validation** | ❌ Untested | 0 | 🔴 HIGH |
| **CORS Headers** | ⚠️ Partial | 1 | 🟡 MEDIUM |
| **Error Handling** | ❌ Untested | 0 | 🟡 MEDIUM |
| **Integration** | ✅ Partial | 10 | 🟢 DONE |
| **Client IP** | ❌ Untested | 0 | 🟢 LOW |
| **Performance** | ❌ Untested | 0 | 🟢 LOW |

---

## Estimated Impact

Once these improvements are implemented:
- **Code Coverage:** 25% → 65%+
- **Bug Detection:** Early identification of edge cases
- **Refactoring Safety:** Confidence in making changes
- **Maintainability:** Easier to understand code behavior
- **Reliability:** Reduced production issues

---

## Quick Start for Contributors

To add tests:

```bash
# Run existing tests
make test

# Run Go tests only (once created)
go test -v ./...

# Generate coverage report
go test -cover ./...

# Run with coverage profile
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

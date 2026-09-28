# 🚀 CORS Proxy - Open Source Edition

A lightning-fast, simple CORS proxy server written in Go. Deploy anywhere with one click!

[![CI](https://github.com/melihbirim/corsproxy/actions/workflows/ci.yml/badge.svg)](https://github.com/melihbirim/corsproxy/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/melihbirim/corsproxy/graph/badge.svg)](https://codecov.io/gh/melihbirim/corsproxy)

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template?template=https://github.com/melihbirim/corsproxy)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

## 🌐 Live Demo

Try it now: **[https://corsproxy-8uo5.onrender.com](https://corsproxy-8uo5.onrender.com)**

**Example:**

```bash
# Fetch GitHub API through the proxy
curl "https://corsproxy-8uo5.onrender.com/?url=https://api.github.com/users/melihbirim"

# Use in JavaScript
fetch('https://corsproxy-8uo5.onrender.com/?url=https://api.example.com/data')
  .then(r => r.json())
  .then(data => console.log(data));
```

## ✨ Features

- ⚡ **Fast**: Written in Go for maximum performance
- 🐳 **Docker Ready**: Full Docker and Docker Compose support
- 🚀 **One-Click Deploy**: Deploy to Railway, Render, Fly.io, or Koyeb
- 🔓 **Full CORS Support**: Handles all CORS headers automatically
- 📦 **Zero Dependencies**: Uses only Go standard library
- 🔒 **Secure**: 10MB request size limit, 30s timeout
- 💾 **Lightweight**: ~10MB Docker image (Alpine-based)

## 🎯 Quick Start

### Local Development

```bash
# Clone the repository
git clone https://github.com/melihbirim/corsproxy.git
cd corsproxy

# Run directly with Go
go run main.go

# Or use Make
make run

# Or build and run
make build
./bin/corsproxy
```

Server starts at `http://localhost:8080`

### Run Tests

```bash
# Make sure server is running in another terminal
make test

# Or run directly
./test.sh
```

### Using Docker

```bash
# Build and run with Docker
docker build -t corsproxy .
docker run -p 8080:8080 corsproxy

# Or use Docker Compose
docker-compose up
```

### Development with Hot Reload

```bash
# Using the development Dockerfile
docker build -f Dockerfile.dev -t corsproxy-dev .
docker run -p 8080:8080 -v $(pwd):/app corsproxy-dev
```

## 📖 Usage

### Basic Request

```bash
curl "http://localhost:8080/?url=https://api.example.com/data"
```

### From JavaScript

```javascript
fetch("http://localhost:8080/?url=https://api.example.com/data")
  .then((response) => response.json())
  .then((data) => console.log(data));
```

### With Custom Headers

```javascript
fetch("http://localhost:8080/?url=https://api.example.com/data", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: "Bearer token123",
  },
  body: JSON.stringify({ key: "value" }),
})
  .then((response) => response.json())
  .then((data) => console.log(data));
```

### Health Check

```bash
curl http://localhost:8080/health
```

Response:

```json
{ "status": "ok", "timestamp": "2026-01-08T12:00:00Z" }
```

## 🌐 One-Click Deployments

### Railway

1. Click the "Deploy on Railway" button above
2. Connect your GitHub repository
3. Railway will auto-detect and deploy
4. Your proxy will be live at `https://your-app.railway.app`

### Render

1. Click the "Deploy to Render" button above
2. Connect your repository
3. Render will build and deploy automatically
4. Access at `https://your-app.onrender.com`

### Fly.io

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login

# Deploy
flyctl launch
```

### Koyeb

```bash
# Install Koyeb CLI
curl -fsSL https://cli.koyeb.com/install.sh | sh

# Login
koyeb login

# Deploy
koyeb app create corsproxy \
  --git github.com/melihbirim/corsproxy \
  --git-branch main \
  --ports 8080:http \
  --routes /:8080
```

## ⚙️ Configuration

### Environment Variables

| Variable                | Default    | Description                                                |
| ----------------------- | ---------- | ---------------------------------------------------------- |
| `PORT`                  | `8080`     | Server port                                                |
| `MAX_REQUEST_SIZE`      | `10485760` | Max request size in bytes (10MB default)                   |
| `REQUEST_TIMEOUT`       | `30s`      | Request timeout (Go duration: 30s, 1m, etc)                |
| `MAX_REDIRECTS`         | `10`       | Maximum number of redirects to follow                      |
| `ALLOWED_ORIGINS`       | `*`        | CORS allowed origins (\* for all, or comma-separated list) |
| `ALLOWED_HOSTS`         | ``         | Comma-separated list of allowed hosts (empty = all)        |
| `BLOCKED_HOSTS`         | ``         | Comma-separated list of blocked hosts                      |
| `ALLOW_PRIVATE_NETWORKS` | `false`   | Allow proxying to loopback/private/link-local addresses (only for trusted internal deployments) |
| `RATE_LIMIT_PER_MINUTE` | `0`        | Rate limit per IP (0 = disabled)                           |
| `REQUIRE_API_KEY`       | `false`    | Require a valid key from `API_KEYS` on every proxied request |
| `API_KEYS`              | ``         | Comma-separated list of accepted API keys (see below)      |
| `DAILY_REQUEST_LIMIT`   | `0`        | Global request ceiling per rolling 24h, across all clients (0 = disabled) |
| `VERBOSE_LOGGING`       | `false`    | Enable detailed request logging                            |

### Production Configuration Example

```bash
# .env file or deployment settings
PORT=8080
MAX_REQUEST_SIZE=5242880              # 5MB
REQUEST_TIMEOUT=15s
RATE_LIMIT_PER_MINUTE=100            # 100 requests per minute per IP
ALLOWED_ORIGINS=https://example.com,https://app.example.com,https://admin.example.com
ALLOWED_HOSTS=api.github.com,api.stripe.com,httpbin.org
BLOCKED_HOSTS=internal.example.com
DAILY_REQUEST_LIMIT=100000
VERBOSE_LOGGING=true
```

### Security Features

**CORS Origins:**

```bash
# Allow all origins (development only)
ALLOWED_ORIGINS=*

# Allow specific origins (production recommended)
ALLOWED_ORIGINS=https://example.com,https://app.example.com
```

**Host Filtering:**

```bash
# Only allow specific APIs
ALLOWED_HOSTS=api.github.com,api.stripe.com

# Block specific hosts (matches the host and its subdomains)
BLOCKED_HOSTS=internal.example.com

# Private, loopback, link-local (incl. cloud metadata 169.254.169.254) and
# CGNAT addresses are always blocked, checked on the resolved IP of every
# connection, including redirects. Set ALLOW_PRIVATE_NETWORKS=true only if
# the proxy must reach internal services.
```

**Rate Limiting:**

```bash
# Limit to 100 requests per minute per IP address
RATE_LIMIT_PER_MINUTE=100
```

**Request Limits:**

```bash
# Smaller file size limit for production
MAX_REQUEST_SIZE=5242880  # 5MB

# Faster timeout for better resource usage
REQUEST_TIMEOUT=15s
```

**API Key Authentication (optional):**

```bash
# Require a key on every proxied request (health checks are unaffected)
REQUIRE_API_KEY=true
API_KEYS=key-for-team-a,key-for-team-b

# Callers send it as a header or query param
curl -H "X-API-Key: key-for-team-a" "https://your-proxy.com/?url=https://api.github.com"
curl "https://your-proxy.com/?url=https://api.github.com&apikey=key-for-team-a"
```

Off by default so the zero-config quick start keeps working; turn it on once the proxy is reachable from outside your own machine.

**Daily Request Ceiling (kill switch):**

```bash
# Hard stop at 100k proxied requests per rolling 24h window, across every
# client combined — a cost/abuse backstop independent of RATE_LIMIT_PER_MINUTE,
# which only limits a single IP.
DAILY_REQUEST_LIMIT=100000
```

This is process-local (an in-memory counter), so it resets on restart and doesn't coordinate across multiple replicas. Good enough as a single-instance cost cap; not a substitute for a shared rate limiter if you run more than one instance.

## ✅ Production Checklist

The zero-config defaults are wide open, by design, so `go run .` and try-it-now demos work with no setup. Before exposing this on the public internet, go through this list:

- [ ] **Set `ALLOWED_ORIGINS`** to your actual site(s). The default `*` lets any website's JavaScript call your proxy.
- [ ] **Set `ALLOWED_HOSTS`** to the specific upstream APIs you intend to proxy to. An empty list lets your proxy reach *any* public host, which makes it attractive as a free open relay for other people's traffic.
- [ ] **Set `RATE_LIMIT_PER_MINUTE`** to bound per-IP abuse.
- [ ] **Set `DAILY_REQUEST_LIMIT`** as a cost backstop if you're paying for egress/compute.
- [ ] **Consider `REQUIRE_API_KEY`** if the proxy is meant for your own apps/services rather than the general public — it turns an open relay into an authenticated one.
- [ ] Leave `ALLOW_PRIVATE_NETWORKS` off unless you specifically need the proxy to reach internal/loopback addresses; it's an SSRF hole otherwise.
- [ ] Run behind HTTPS (a platform load balancer, or a reverse proxy like Caddy/nginx) — this server speaks plain HTTP.
- [ ] Watch the startup logs: a `⚠️  Running with open defaults` line lists exactly which of the above are still off.

## 🏗️ Project Structure

```bash
corsproxy/
├── main.go              # Main application
├── go.mod              # Go module file
├── Makefile            # Build automation
├── test.sh             # Test script
├── Dockerfile          # Production Docker image
├── Dockerfile.dev      # Development Docker image
├── docker-compose.yml  # Docker Compose config
├── railway.json        # Railway deployment config
├── render.yaml         # Render deployment config
├── fly.toml           # Fly.io deployment config
├── koyeb.json         # Koyeb deployment config
├── .env.example       # Environment variables template
└── README.md          # This file
```

## 🛠️ Makefile Commands

```bash
make help          # Show all available commands
make build         # Build the binary
make run           # Run the server
make test          # Run tests
make docker-build  # Build Docker image
make docker-run    # Run Docker container
make clean         # Clean build artifacts
make fmt           # Format code
make lint          # Lint code
```

## 🔧 API Endpoints

### `GET/POST/PUT/DELETE/PATCH /?url=<target-url>`

Proxies the request to the target URL with CORS headers.

**Query Parameters:**

- `url` (required): The target URL to proxy

**Example:**

```bash
curl "http://localhost:8080/?url=https://api.github.com/users/octocat"
```

### `GET /health`

Health check endpoint for monitoring.

**Response:**

```json
{ "status": "ok", "timestamp": "2026-01-08T12:00:00Z" }
```

## 🛡️ Security Features

- URL validation (must start with http:// or https://)
- Request size limiting (10MB max)
- Request timeout (30s)
- Redirect limit (max 10 redirects)
- No arbitrary code execution

## 🚀 Performance

- **Cold Start**: < 100ms
- **Request Latency**: < 50ms overhead
- **Memory Usage**: ~10MB base
- **Concurrent Requests**: Thousands (Go's goroutines)

## 📝 License

MIT License - feel free to use this in your projects!

## 🤝 Contributing

We welcome contributions! This project is perfect for learning Go and building production-ready software.

### Good First Issues

Looking to contribute? Check out our [Good First Issues](https://github.com/melihbirim/corsproxy/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) - production enhancements perfect for first-time contributors:

**Priority 1 (Great for beginners):**

- Add graceful shutdown handling
- Implement request ID middleware
- Add structured JSON logging
- Add security headers (CSP, HSTS, etc)

**Priority 2 (Intermediate):**

- Add Prometheus metrics endpoint
- Implement response caching
- Add API key authentication
- Add circuit breaker pattern

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Quick Start for Contributors

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes (`git commit -m 'Add: amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📊 Comparison with Other Solutions

| Feature          | This Proxy | CORS Anywhere | corsproxy-node |
| ---------------- | ---------- | ------------- | --------------- |
| Language         | Go         | Node.js       | Node.js         |
| Docker Support   | ✅         | ⚠️            | ✅              |
| One-Click Deploy | ✅         | ❌            | ⚠️              |
| Memory Usage     | ~10MB      | ~50MB         | ~40MB           |
| Cold Start       | <100ms     | ~1s           | ~800ms          |
| Dependencies     | 0          | Many          | Many            |

## 💡 Use Cases

- Bypass CORS restrictions in development
- Access APIs that don't support CORS
- Build web applications that need to fetch external resources
- Create API gateways with CORS support
- Testing and prototyping

## ⚠️ Production Considerations

While this proxy is production-ready, consider:

- **Rate Limiting**: Add rate limiting for public deployments
- **Authentication**: Add API keys if needed
- **Monitoring**: Use the `/health` endpoint for uptime monitoring
- **Logging**: Logs are written to stdout (Docker-friendly)
- **Caching**: Consider adding caching for frequently accessed resources

## 🐛 Troubleshooting

### Port already in use

```bash
# Change port via environment variable
PORT=3000 go run main.go
```

### Docker build fails

```bash
# Clean Docker cache
docker builder prune
docker build --no-cache -t corsproxy .
```

### Connection timeout

The proxy has a 30-second timeout. For longer requests, modify the `client.Timeout` in `main.go`.

## 📞 Support

- Open an issue on GitHub
- Check existing issues for solutions
- Read the code - it's simple and well-commented!

---

Made with ❤️ by [@melihbirim](https://github.com/melihbirim)

**Star ⭐ this repository if you find it useful!**

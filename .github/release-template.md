# 🚀 CORS Proxy v1.0.0

A lightning-fast, simple CORS proxy server written in Go. Deploy anywhere with one click!

## 🌐 Live Demo

Try it now: **https://corsproxy-8uo5.onrender.com**

## ✨ What's New

This is the initial release of CORS Proxy with production-ready features:

### Core Features

- ⚡ **Fast Go-based proxy** - Maximum performance with minimal overhead
- 🔓 **Full CORS support** - Handles all CORS headers automatically
- 🎯 **Zero dependencies** - Uses only Go standard library
- 🐳 **Docker ready** - Multi-stage builds for minimal image size (~10MB)

### Configuration & Security

- 🔒 **Rate limiting** - Configurable per-IP rate limits
- 🛡️ **Host filtering** - Allowlist/blocklist support
- ⚙️ **Environment-based config** - Easy deployment configuration
- 🔐 **Security limits** - Request size (10MB) and timeout (30s) protection

### Deployment

- 🚀 **One-click deploy** - Railway, Render, Fly.io, Koyeb support
- 📦 **Docker Compose** - Local development made easy
- 🏥 **Health checks** - Built-in `/health` endpoint

### Developer Experience

- 🧪 **Comprehensive tests** - Automated test suite with 10 test cases
- 🛠️ **Makefile** - Simple build, test, and deploy commands
- 📝 **Documentation** - Complete README with examples
- 🤝 **Contribution ready** - Issue templates and guidelines

## 📥 Installation

### Quick Start

```bash
# Run with Docker
docker run -p 8080:8080 ghcr.io/melihbirim/corsproxy:latest

# Or with Go
go install github.com/melihbirim/corsproxy@latest
```

### From Source

```bash
git clone https://github.com/melihbirim/corsproxy.git
cd corsproxy
make build
./bin/cors-proxy
```

## 📖 Usage

```bash
# Basic usage
curl "http://localhost:8080/?url=https://api.github.com/users/octocat"

# With JavaScript
fetch('http://localhost:8080/?url=https://api.example.com/data')
  .then(r => r.json())
  .then(data => console.log(data));
```

## 🔧 Configuration

Configure via environment variables:

- `PORT` - Server port (default: 8080)
- `RATE_LIMIT_PER_MINUTE` - Rate limit per IP (default: 0/disabled)
- `ALLOWED_ORIGINS` - Comma-separated CORS origins (default: \*)
- `MAX_REQUEST_SIZE` - Max request size in bytes (default: 10MB)
- See [README](https://github.com/melihbirim/corsproxy#configuration) for full list

## 🤝 Contributing

We welcome contributions! Check out our [good first issues](https://github.com/melihbirim/corsproxy/labels/good%20first%20issue) to get started.

## 📊 What's Next

See our [roadmap](https://github.com/melihbirim/corsproxy/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement) for planned features:

- Prometheus metrics endpoint
- Response caching
- API key authentication
- Circuit breaker pattern
- And more!

## 🙏 Acknowledgments

Built with Go and ❤️ by [@melihbirim](https://github.com/melihbirim)

---

**Full Changelog**: https://github.com/melihbirim/corsproxy/blob/main/CHANGELOG.md

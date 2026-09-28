# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.2] - 2026-09-28

### Added

- Test coverage raised from 43.7% to 84.0% (env-parsing helpers, config loading, rate limiting, redirect handling, and more of `corsProxyHandler`'s branches).
- `.golangci.yml` (errcheck, govet, staticcheck, unused); `make lint` is now enforced against a real config instead of whatever defaults happen to be installed.

## [1.2.1] - 2026-09-28

### Fixed

- `go.mod` module path corrected from `github.com/melihbirim/cors-proxy` to `github.com/melihbirim/corsproxy`, matching the actual repo. The mismatch broke `go install`/pkg.go.dev indexing for anyone using this as a module.

## [1.2.0] - 2026-09-28

### Added

- Optional API key authentication (`REQUIRE_API_KEY`, `API_KEYS`), checked with a constant-time comparison. Off by default so the zero-config quick start keeps working.
- Global daily request ceiling / kill switch (`DAILY_REQUEST_LIMIT`), independent of the per-IP rate limiter.
- Startup warning when `RATE_LIMIT_PER_MINUTE`, `ALLOWED_HOSTS`, `ALLOWED_ORIGINS`, and `REQUIRE_API_KEY` are all left at their wide-open defaults.
- README "Production Checklist" section.

## [1.1.0] - 2026-09-27

### Security

- Block requests to private, loopback, link-local (including cloud metadata at 169.254.169.254), CGNAT, and private IPv6 addresses by default. The check runs on the resolved IP of every connection, so it also covers redirects and DNS rebinding. Opt out with `ALLOW_PRIVATE_NETWORKS=true`.
- `ALLOWED_HOSTS` / `BLOCKED_HOSTS` now match the exact host or its subdomains, parsed with `net/url` (previously substring matching, which `user@host` URLs and look-alike domains could bypass).
- Redirect targets are re-checked against the host rules.
- Upstream `Set-Cookie` headers are dropped, and proxied responses are sent with `Content-Security-Policy: sandbox` and `X-Content-Type-Options: nosniff`.
- Request bodies are capped at `MAX_REQUEST_SIZE`; oversized upstream responses return 502 instead of being truncated.
- The rate limiter uses the rightmost `X-Forwarded-For` entry (set by the platform proxy) instead of the client-controlled leftmost one, and expired entries are pruned.

## [1.0.0] - 2026-01-08

### Added

- Initial release of CORS Proxy
- Fast Go-based CORS proxy server
- Full CORS support with configurable origins
- Environment-based configuration
- Rate limiting (per IP, configurable)
- Host allowlist/blocklist filtering
- Request size limits (configurable, default 10MB)
- Request timeout (configurable, default 30s)
- Health check endpoint (`/health`)
- Docker support with multi-stage builds
- Docker Compose configuration
- One-click deployment configs for Railway, Render, Fly.io, Koyeb
- Makefile for easy building and testing
- Comprehensive test suite (test.sh)
- Automated code formatting and linting
- Graceful error handling
- Verbose logging option
- Zero external dependencies

### Infrastructure

- GitHub issue templates (bug report, feature request, good first issue)
- Pull request template
- Contributing guidelines (CONTRIBUTING.md)
- MIT License
- Comprehensive README with examples
- Production-ready configuration examples

### Security

- URL validation
- Request size limiting
- Timeout protection
- Rate limiting support
- Host filtering (allow/block lists)
- Proper CORS implementation with credentials support

[1.0.0]: https://github.com/melihbirim/corsproxy/releases/tag/v1.0.0

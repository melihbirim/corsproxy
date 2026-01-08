# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

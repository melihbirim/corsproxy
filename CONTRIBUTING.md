# Contributing to CORS Proxy

Thank you for your interest in contributing! 🎉

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/cors-proxy.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit your changes: `git commit -m "Add: your feature description"`
8. Push to your fork: `git push origin feature/your-feature-name`
9. Open a Pull Request

## Development Setup

```bash
# Install Go 1.21 or later
# Clone the repository
git clone https://github.com/melihbirim/cors-proxy.git
cd cors-proxy

# Install dependencies
make install

# Run the server
make run

# Run tests
make test

# Run linter
make lint
```

## Code Style

- We use `gofmt` for formatting (runs automatically with `make build` and `make lint`)
- Follow Go best practices and idioms
- Keep functions small and focused
- Add comments for exported functions
- Handle errors explicitly

## Testing

- All new features should include tests
- Update `test.sh` if adding new endpoints
- Ensure all tests pass before submitting PR
- Test with: `make test`

## Pull Request Process

1. **Update documentation** - Update README.md if adding features
2. **Add tests** - Ensure your code is tested
3. **Run linter** - `make lint` should pass with 0 issues
4. **One feature per PR** - Keep PRs focused on a single change
5. **Clear commit messages** - Use descriptive commit messages
6. **Reference issues** - Link to relevant issues in your PR description

## Commit Message Format

```bash
Type: Brief description

Longer description if needed

Fixes #123
```

Types:

- `Add:` New feature
- `Fix:` Bug fix
- `Refactor:` Code refactoring
- `Docs:` Documentation changes
- `Test:` Test additions or changes
- `Chore:` Maintenance tasks

Examples:

- `Add: Graceful shutdown support`
- `Fix: Rate limiter memory leak`
- `Docs: Update environment variables section`

## Good First Issues

Looking for something to work on? Check out issues labeled:

- `good first issue` - Great for newcomers
- `help wanted` - We'd love help with these
- `enhancement` - New features to implement

## Areas for Contribution

### Priority 1 (Good First Issues)

- [ ] Graceful shutdown handling
- [ ] Request ID middleware
- [ ] Structured JSON logging
- [ ] Security headers (CSP, X-Frame-Options, HSTS)

### Priority 2 (Intermediate)

- [ ] Prometheus metrics endpoint
- [ ] Response caching with TTL
- [ ] API key authentication
- [ ] Circuit breaker for failing backends

### Priority 3 (Advanced)

- [ ] Redis-backed rate limiting
- [ ] Distributed tracing integration
- [ ] Request/response body logging (debug mode)
- [ ] WebSocket proxy support

## Questions?

- Open an issue for bug reports or feature requests
- Start a discussion for questions or ideas
- Check existing issues before creating new ones

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing!** 🙏
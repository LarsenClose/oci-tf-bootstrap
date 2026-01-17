# Contributing to oci-tf-bootstrap

Thank you for your interest in contributing!

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/larsenclose/oci-tf-bootstrap.git
   cd oci-tf-bootstrap
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   go test -v ./...
   ```

4. Build:
   ```bash
   make build
   ```

## Running Locally

You need OCI CLI configured (`~/.oci/config`) with valid credentials to run the tool:

```bash
./oci-tf-bootstrap --output ./terraform-test
```

## Code Quality

Before submitting a PR, ensure:

1. **Tests pass:**
   ```bash
   go test -v -race ./...
   ```

2. **Linting passes:**
   ```bash
   golangci-lint run
   ```

3. **Code is formatted:**
   ```bash
   go fmt ./...
   ```

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Commit with clear messages
7. Push to your fork
8. Open a Pull Request

## Commit Messages

Use clear, descriptive commit messages:

- `feat: add support for OKE cluster discovery`
- `fix: handle nil pointer in shape discovery`
- `docs: update always-free tier documentation`
- `test: add tests for image filtering`

## Testing

- Unit tests should not require OCI credentials
- Use mocks for OCI SDK client interfaces
- Integration tests (if any) should be clearly marked and optional

## Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Add comments for non-obvious logic
- Keep functions focused and small

## Questions?

Open an issue for discussion before starting major changes.

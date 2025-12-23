# Contributing to Modbus Proxy Manager

Thank you for your interest in contributing to Modbus Proxy Manager! This document provides guidelines for contributing to the project.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/modbus-proxy-manager.git
   cd modbus-proxy-manager
   ```
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)

### Installing Dependencies

```bash
go mod download
```

### Running the Application

```bash
make run
# or
go run main.go
```

## Making Changes

### Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused

### Testing

- Write tests for new functionality
- Ensure all tests pass before submitting:
  ```bash
  make test
  # or
  go test -v ./...
  ```
- Aim for good test coverage

### Commits

- Write clear, concise commit messages
- Use present tense ("Add feature" not "Added feature")
- Reference issues in commits when applicable
- Keep commits atomic and focused

Example commit message:
```
Add health check endpoint

- Implements /api/health endpoint
- Returns JSON status response
- Useful for Docker healthchecks
```

## Submitting Changes

1. **Push your changes** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub:
   - Provide a clear description of the changes
   - Reference any related issues
   - Ensure CI checks pass

3. **Address review feedback** if requested

## Pull Request Guidelines

- One feature/fix per pull request
- Include tests for new functionality
- Update documentation if needed
- Ensure the code builds and all tests pass
- Keep pull requests focused and reasonably sized

## Code Review Process

1. A maintainer will review your pull request
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR

## Reporting Bugs

- Use the GitHub issue tracker
- Provide a clear description of the bug
- Include steps to reproduce
- Mention your environment (OS, Go version, etc.)
- Include relevant logs or error messages

## Suggesting Features

- Use the GitHub issue tracker
- Clearly describe the feature and its benefits
- Explain the use case
- Be open to discussion and feedback

## Development Workflow

### Building

```bash
make build
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run linter
make lint
```

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
make docker-run
```

## Project Structure

```
.
├── main.go              # Application entry point
├── pkg/
│   ├── api/            # HTTP API handlers
│   ├── auth/           # Authentication logic
│   ├── config/         # Configuration management
│   ├── logger/         # Logging system
│   ├── manager/        # Proxy manager
│   ├── modbus/         # Modbus protocol handling
│   ├── proxy/          # Proxy instance logic
│   └── web/            # Web UI assets
├── .github/
│   └── workflows/      # CI/CD workflows
└── Dockerfile          # Container build file
```

## Questions?

If you have questions, feel free to:
- Open an issue on GitHub
- Ask in the pull request
- Check existing documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to Modbus Proxy Manager!

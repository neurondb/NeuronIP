# Contributing to NeuronIP

Thank you for your interest in contributing to NeuronIP! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- PostgreSQL 16+ with NeuronDB extension
- Go 1.24+ (for backend development)
- Node.js 18+ (for frontend development)

### Setting Up the Development Environment

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/NeuronIP.git
   cd NeuronIP
   ```

2. **Set up environment variables**
   ```bash
   # Copy example environment files
   cp .env.example .env
   cp frontend/.env.example frontend/.env.local
   ```

3. **Start services with Docker Compose**
   ```bash
   docker compose up -d
   ```

4. **For local development**

   **Backend:**
   ```bash
   cd api
   go mod download
   go run cmd/server/main.go
   ```

   **Frontend:**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Development Workflow

### Branch Naming

- Feature branches: `feature/description-of-feature`
- Bug fixes: `fix/description-of-bug`
- Documentation: `docs/description-of-docs`
- Refactoring: `refactor/description-of-refactor`

### Making Changes

1. **Create a branch** from `main` or `develop`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards below

3. **Test your changes**
   - Run backend tests: `cd api && go test ./...`
   - Run frontend tests: `cd frontend && npm test`
   - Run linters: `npm run lint` (frontend) or `golangci-lint run` (backend)

4. **Commit your changes** with clear commit messages
   ```bash
   git commit -m "feat: add new feature description"
   ```

5. **Push and create a Pull Request**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, missing semicolons, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Example:
```
feat(semantic): add vector similarity search
fix(warehouse): resolve SQL query timeout issue
docs: update API documentation
```

## Coding Standards

### Go (Backend)

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` or `goimports` for formatting
- Run `golangci-lint` before committing
- Write tests for new functionality
- Document exported functions with comments

### TypeScript/React (Frontend)

- Follow the existing code style in the project
- Use TypeScript for type safety
- Follow React best practices and hooks patterns
- Use functional components
- Run `npm run lint` before committing
- Write tests for components and utilities

### General

- Keep functions small and focused
- Write clear, descriptive variable and function names
- Add comments for complex logic
- Remove debug code before committing

## Testing

### Backend Testing

```bash
cd api
go test ./...
go test -v -coverprofile=coverage.out ./...
```

### Frontend Testing

```bash
cd frontend
npm test
npm test -- --coverage
```

### Integration Testing

Ensure Docker Compose services are running and test the full stack:
```bash
docker compose up -d
# Run integration tests
```

## Pull Request Process

1. **Ensure your PR is ready**
   - All tests pass
   - Code is properly formatted
   - Documentation is updated if needed
   - No merge conflicts with the base branch

2. **Fill out the PR template** completely

3. **Request review** from maintainers

4. **Address feedback** from reviewers

5. **Wait for approval** before merging

## Documentation

- Update README.md if you change setup or usage instructions
- Add or update API documentation if you modify endpoints
- Update CHANGELOG.md for user-facing changes
- Comment complex code and algorithms

## Questions?

- Open a discussion for general questions
- Open an issue for bugs or feature requests
- Check existing issues and PRs before creating new ones

Thank you for contributing to NeuronIP! 🎉

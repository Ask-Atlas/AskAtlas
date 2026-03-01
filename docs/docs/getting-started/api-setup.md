---
sidebar_position: 2
---

# API Setup

The backend API for AskAtlas, built with [Go](https://go.dev/).

## Prerequisites

- [Go](https://go.dev/dl/) — Version 1.24 or higher
- [Docker](https://www.docker.com/) — For containerization
- [Infisical CLI](https://infisical.com/docs/cli/overview) — For environment variables

## Dependencies

- [golangci-lint](https://golangci-lint.run/) — Code quality
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) — Import formatting
- [chi](https://github.com/go-chi/chi) — REST API framework

## Getting Started

1. **Install dependencies:**

   ```bash
   make install
   ```

2. **Run the development server:**
   ```bash
   make dev
   ```

## Makefile Commands

### Development & Build

| Command | Description |
|---------|-------------|
| `make dev` | Start the development server (`go run`). |
| `make stage` | Start the staging server (`go run`). |
| `make prod` | Start the production server (`go run`). |
| `make build` | Build the binary to `api`. |

### Code Quality

| Command | Description |
|---------|-------------|
| `make lint` | Run `golangci-lint`. |
| `make format` | Format code using `goimports -w`. |
| `make format-check` | Check formatting using `goimports -l`. |
| `make tidy` | Run `go mod tidy` and `go mod verify`. |
| `make tidy-check` | Check if modules are tidy via script. |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run unit tests. |
| `make mockery` | Generate mocks for interfaces. |

### Docker

| Command | Description |
|---------|-------------|
| `make docker-build` | Build the Docker image `askatlas-api`. |
| `make docker-run` | Run the container on port 8080 with Infisical secrets. |

## Project Structure

```
api/
├── cmd/          # Application entrypoints
├── internal/     # Private application code and business logic
│   ├── clerk/    # Clerk webhook event processing
│   ├── db/       # sqlc-generated database code
│   ├── files/    # Files domain (model, service, repository, handler)
│   ├── handlers/ # HTTP handlers
│   ├── middleware/ # Auth and request middleware
│   ├── user/     # User domain
│   └── utils/    # Shared utilities
├── pkg/          # Public packages (apperrors, authctx)
├── db/queries/   # SQL query files for sqlc
├── e2e/          # End-to-end tests
└── scripts/      # Helper scripts
```

# AskAtlas API

The backend API for AskAtlas, built with [Go](https://go.dev/).

## Prerequisites

Before getting started, ensure you have the following installed:

- [Go](https://go.dev/dl/) - Version 1.24 or higher
- [Docker](https://www.docker.com/) - For containerization
- [Infisical CLI](https://infisical.com/docs/cli/overview) - For environment variables

## Dependencies Installed
- [golangci-lint](https://golangci-lint.run/) - Code Quality
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) - Import Formatting
- [chi](https://github.com/go-chi/chi) - REST API Framework


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

This project uses a `makefile` to standardize common tasks. Run `make <command>` to execute them.

### Development & Build

| Command             | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `make dev`          | Starts the development server (`go run`).            |
| `make stage`        | Starts the development server (`go run`).            |
| `make prod`         | Starts the development server (`go run`).            |
| `make build`        | Builds the binary to `api`.                          |

### Code Quality

| Command             | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `make lint`         | Runs `golangci-lint`.                                |
| `make format`       | Formats code using `goimports -w`.                   |
| `make format-check` | Checks formatting using `goimports -l`.              |
| `make tidy`         | Runs `go mod tidy` and `go mod verify`.              |
| `make tidy-check`   | Checks if modules are tidy via script.               |

### Testing

| Command     | Description      |
| ----------- | ---------------- |
| `make test` | Runs unit tests. |

### Docker

| Command             | Description                                                                     |
| ------------------- | ------------------------------------------------------------------------------- |
| `make docker-build` | Builds the Docker image `askatlas-api`.                                         |
| `make docker-run`   | Runs the Docker container on port 8080 with Infisical secrets from `.env.local` |

## Project Structure

- `cmd/`: Application entrypoints.
- `internal/`: Private application code and business logic.
- `scripts/`: Helper scripts for build and maintenance.
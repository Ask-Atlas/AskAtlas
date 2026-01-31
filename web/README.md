# AskAtlas Web Client

The web client for AskAtlas, built with [Next.js](https://nextjs.org).

## Prerequisites

Before getting started, ensure you have the following installed:

- [Node.js](https://nodejs.org/) - Version 22 or higher
- [pnpm](https://pnpm.io/) - For package management
- [Docker](https://www.docker.com/) - For containerization
- [Infisical CLI](https://infisical.com/docs/cli/overview) - For environment variables

## Getting Started

1. **Install dependencies:**

   ```bash
   make install
   ```

2. **Run the development server:**
   ```bash
   make dev
   ```
   This will start the app on [http://localhost:3000](http://localhost:3000) with dev environment variables loaded via Infisical.

## Makefile Commands

This project uses a `makefile` to standardize common tasks. Run `make <command>` to execute them.

### Development & Build

| Command        | Description                                                 |
| -------------- | ----------------------------------------------------------- |
| `make dev`     | Starts the development server using Infisical for env vars. |
| `make build`   | Builds the application for production.                      |
| `make staging` | Builds and runs the app in staging mode.                    |
| `make prod`    | Builds and runs the app in production mode.                 |
| `make clean`   | Removes `.next` and `node_modules`.                         |

### Code Quality

| Command             | Description                            |
| ------------------- | -------------------------------------- |
| `make lint`         | Runs the linter.                       |
| `make lint-fix`     | Runs the linter and fixes issues.      |
| `make format`       | Formats code using Prettier.           |
| `make format-check` | Checks if code is correctly formatted. |
| `make typecheck`    | Runs TypeScript type checking.         |

### Testing

| Command                 | Description                                  |
| ----------------------- | -------------------------------------------- |
| `make test`             | Runs unit/integration tests.                 |
| `make test-e2e`         | Runs Playwright E2E tests (dev environment). |
| `make test-e2e-staging` | Runs E2E tests against staging.              |
| `make test-e2e-prod`    | Runs E2E tests against production.           |
| `make test-e2e-codegen` | Opens Playwright codegen for creating tests. |
| `make e2e-report`       | Shows the Playwright test report.            |

### Docker

| Command                 | Description                                     |
| ----------------------- | ----------------------------------------------- |
| `make docker-build`     | Builds the Docker image `askatlas-web`.         |
| `make docker-run-local` | Runs the Docker container locally on port 3000. |

## Project Structure

- `app/`: Next.js App Router pages and layouts.
- `lib/`: Shared utilities and business logic.
- `public/`: Static assets.
- `e2e/`: Playwright end-to-end tests.

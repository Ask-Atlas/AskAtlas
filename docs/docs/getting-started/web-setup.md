---
sidebar_position: 3
---

# Web Setup

The web client for AskAtlas, built with [Next.js](https://nextjs.org) (App Router).

## Prerequisites

- [Node.js](https://nodejs.org/) — Version 22 or higher
- [pnpm](https://pnpm.io/) — For package management
- [Docker](https://www.docker.com/) — For containerization
- [Infisical CLI](https://infisical.com/docs/cli/overview) — For environment variables

## Getting Started

1. **Install dependencies:**

   ```bash
   make install
   ```

2. **Run the development server:**
   ```bash
   make dev
   ```
   This starts the app on [http://localhost:3000](http://localhost:3000) with dev environment variables loaded via Infisical.

## Makefile Commands

### Development & Build

| Command | Description |
|---------|-------------|
| `make dev` | Start the dev server using Infisical for env vars. |
| `make build` | Build the application for production. |
| `make staging` | Build and run the app in staging mode. |
| `make prod` | Build and run the app in production mode. |
| `make clean` | Remove `.next` and `node_modules`. |

### Code Quality

| Command | Description |
|---------|-------------|
| `make lint` | Run the linter. |
| `make lint-fix` | Run the linter and fix issues. |
| `make format` | Format code using Prettier. |
| `make format-check` | Check if code is correctly formatted. |
| `make typecheck` | Run TypeScript type checking. |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | Run unit/integration tests. |
| `make test-e2e` | Run Playwright E2E tests (dev). |
| `make test-e2e-staging` | Run E2E tests against staging. |
| `make test-e2e-prod` | Run E2E tests against production. |
| `make test-e2e-codegen` | Open Playwright codegen for creating tests. |
| `make e2e-report` | Show the Playwright test report. |

### Docker

| Command | Description |
|---------|-------------|
| `make docker-build` | Build the Docker image `askatlas-web`. |
| `make docker-run-local` | Run the container locally on port 3000. |

## Project Structure

```
web/
├── app/              # Next.js App Router pages and layouts
│   ├── (dashboard)/  # Authenticated dashboard pages
│   ├── (marketing)/  # Public marketing/landing pages
│   └── practice/     # Practice page
├── components/       # Shared UI components (shadcn/ui, animate-ui)
├── lib/              # Shared utilities and business logic
│   └── features/     # Feature-based code organization
│       ├── dashboard/ # Dashboard-specific components and logic
│       └── marketing/ # Marketing-specific components and logic
├── hooks/            # Custom React hooks
├── public/           # Static assets
└── e2e/              # Playwright end-to-end tests
```

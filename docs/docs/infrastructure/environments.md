---
sidebar_position: 3
---

# Environments

AskAtlas runs three environments: **dev**, **stage**, and **prod**. Each has isolated containers, ports, and secrets.

## Environment Overview

Each environment has isolated containers, ports, memory limits, and secrets. Port mappings and resource limits are defined in `scripts/deploy-common.sh`. Containers follow the `<app>-<env>` naming convention.

## Required Secrets

Secrets are managed through [Infisical](https://infisical.com) and injected at container startup. Each service requires its own set of credentials (database connection, authentication keys, etc.).

For the full list of required environment variables, check the Infisical project for each service. Deployment credentials (SSH, Infisical machine identities) are stored as GitHub Actions secrets per environment.

## Infisical Setup

### Local Development

The Infisical CLI is used to inject secrets during local development. The `make dev` targets in both `api/` and `web/` use `infisical run` to wrap the dev server.

Ensure you have:
1. Installed the [Infisical CLI](https://infisical.com/docs/cli/overview)
2. Logged in via `infisical login`
3. Configured `.infisical.json` in the service directory (already committed)

### Production Containers

Secrets are injected at container startup via Infisical's universal-auth method. The container receives Infisical credentials as environment variables, authenticates at startup in `start.sh`, and runs the application with all secrets loaded.

See [Deployment](./deployment) for details on the secret injection flow.

## Running the Full Stack Locally

1. **Database** — Ensure PostgreSQL is running and accessible. Run migrations from `migrations/`:
   ```bash
   ENV=dev make migrate-up
   ```

2. **API** — From `api/`:
   ```bash
   make install
   make dev
   ```

3. **Web** — From `web/`:
   ```bash
   make install
   make dev
   ```

The web app runs on `localhost:3000` and the API on `localhost:8080`.

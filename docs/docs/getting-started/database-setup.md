---
sidebar_position: 4
---

# Database Setup

## Prerequisites

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Infisical CLI](https://infisical.com/)

## Makefile Commands

All commands must be run from the `migrations/` directory.

| Command | Description | Example |
|---------|-------------|---------|
| `make migrate-up` | Migrate database up to latest version. | `ENV=dev make migrate-up` |
| `make migrate-down` | Migrate database down. Requires `version` or `steps`. | `ENV=dev make migrate-down version=20240101120000` or `ENV=dev make migrate-down steps=1` |
| `make migrate-status` | Show migration status. | `ENV=dev make migrate-status` |
| `make migrate-create` | Create a new migration. Requires `name`. | `ENV=dev make migrate-create name=add_users_table` |

## Folder Structure

```
migrations/
└── sql/    # SQL migration files (up and down)
```

See the [Database Migrations](../development/database-migrations) guide for conventions and workflow details.

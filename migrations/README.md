# Migrations

## Prerequisites

- [migrate](https://github.com/golang-migrate/migrate)
- [infisical](https://infisical.com/)

## Makefile Commands

This project uses a `makefile` to standardize common tasks. Run `make <command>` to execute them. These commands must be run from the migrations directory.

### Migrations

| Command               | Description                                                                                         | Example usage                                          |
| --------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| `make migrate-up`     | Migrates the database up to the latest version for the given environment.                          | `ENV=dev make migrate-up`                              |
| `make migrate-down`   | Migrates the database down for the given environment. Requires either a `version` or `steps` arg.  | `ENV=dev make migrate-down version=20240101120000` or `ENV=dev make migrate-down steps=1` |
| `make migrate-status` | Shows the migration status for the given environment.                                              | `ENV=dev make migrate-status`                          |
| `make migrate-create` | Creates a new migration for the given environment. Requires a `name` arg for the migration.        | `ENV=dev make migrate-create name=add_users_table`     |


## Folder Structure

- `sql/`: SQL migration files.
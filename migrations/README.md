# Migrations

## Prerequisites

- [migrate](https://github.com/golang-migrate/migrate)
- [infisical](https://infisical.com/)

## Makefile Commands

This project uses a `makefile` to standardize common tasks. Run `make <command>` to execute them. These commands must be run from the migrations directory.

### Migrations

| Command             | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `make migrate-up`   | Migrates up                                            |
| `make migrate-down` | Migrates down                                        |
| `make migrate-status` | Migrates status                                        |
| `make migrate-create` | Migrates create                                        |


## Folder Structure

- `sql/`: SQL migration files.
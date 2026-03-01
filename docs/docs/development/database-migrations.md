---
sidebar_position: 4
---

# Database Migrations

We use [golang-migrate](https://github.com/golang-migrate/migrate) for version-controlled schema changes.

## Naming Convention

Migration files use a **timestamp prefix**:

```
YYYYMMDDHHMMSS_description.up.sql
YYYYMMDDHHMMSS_description.down.sql
```

Example:
```
20260217090948_create_files_table.up.sql
20260217090948_create_files_table.down.sql
```

## Creating a Migration

```bash
ENV=dev make migrate-create name=add_courses_table
```

This creates two files in `migrations/sql/`:
- `<timestamp>_add_courses_table.up.sql` — the forward migration
- `<timestamp>_add_courses_table.down.sql` — the rollback

## Applying Migrations

```bash
# Migrate up to latest
ENV=dev make migrate-up

# Check current status
ENV=dev make migrate-status
```

## Rolling Back

```bash
# Roll back to a specific version
ENV=dev make migrate-down version=20260217090948

# Roll back N steps
ENV=dev make migrate-down steps=1
```

## Writing Guidelines

### Up Migration

- Create tables, indexes, and types
- Use `IF NOT EXISTS` for extensions and types when possible
- Always include `created_at` and `updated_at` with defaults
- Add indexes for common query patterns
- Add comments explaining index purposes

### Down Migration

- Reverse everything the up migration did
- Drop indexes before tables
- Drop types after tables that use them
- Use `IF EXISTS` to make rollbacks idempotent

## After Schema Changes: sqlc

After modifying the database schema, regenerate the sqlc code:

1. Update or add queries in `api/db/queries/<domain>.sql`
2. Run sqlc generate (configured in `api/sqlc.yml`)
3. The generated Go code is placed in `api/internal/db/`
4. Update repository adapters if the generated types changed

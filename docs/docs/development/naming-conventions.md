---
sidebar_position: 3
---

# Naming Conventions

Consistent naming across the stack makes the codebase predictable for all team members.

## API Conventions

### URL Paths

- **kebab-case**, plural nouns
- Use `/me` as a shorthand for the authenticated user

| ✅ Good | ❌ Bad |
|---------|--------|
| `/api/me/files` | `/api/myFiles` |
| `/api/files/:file_id` | `/api/file/:fileId` |
| `/api/study-guides` | `/api/studyGuides` |

### JSON Response Fields

- **`snake_case`** — matches Go struct JSON tags

```json
{
  "id": "...",
  "mime_type": "image/png",
  "created_at": "2026-01-15T00:00:00Z",
  "last_viewed_at": "2026-02-01T12:00:00Z"
}
```

### Query Parameters

- **`snake_case`**

| Parameter | Example |
|-----------|---------|
| `sort_by` | `?sort_by=created_at` |
| `sort_dir` | `?sort_dir=desc` |
| `page_limit` | `?page_limit=25` |
| `min_size` | `?min_size=1024` |
| `created_from` | `?created_from=2026-01-01T00:00:00Z` |

### Error Responses

Structured `AppError` format:

```json
{
  "code": 400,
  "status": "Bad Request",
  "message": "Invalid query parameters",
  "details": {
    "sort_by": "must be one of: updated_at, created_at, name, size, status, mime_type"
  }
}
```

### Go Code

| Element | Convention | Example |
|---------|-----------|---------|
| Package names | Short, lowercase | `files`, `apperrors`, `authctx` |
| Interfaces | Named by behavior | `FileService`, `Repository` |
| Struct types | PascalCase | `FileHandler`, `ListFilesParams` |
| JSON tags | `snake_case` | `` `json:"mime_type"` `` |
| Error variables | `Err` prefix | `ErrNotFound`, `ErrConflict` |

## Frontend Conventions

### Route Groups

- Parenthesized names for layout boundaries

| Group | Purpose |
|-------|---------|
| `(dashboard)` | Authenticated pages with sidebar |
| `(marketing)` | Public pages with navbar + footer |

### Feature Directories

- **kebab-case** under `lib/features/`

```
lib/features/
├── marketing/
│   └── landing/
└── dashboard/
```

### Components

| Element | Convention | Example |
|---------|-----------|---------|
| Component files | kebab-case | `marketing-navbar.tsx`, `auth-buttons.tsx` |
| Component exports | PascalCase | `MarketingNavbar`, `AuthButtons` |
| Page files | `page.tsx` | `app/(dashboard)/home/page.tsx` |
| Layout files | `layout.tsx` | `app/(dashboard)/layout.tsx` |

### i18n

Dictionaries and providers organized per feature:

```
lib/features/<feature>/i18n/
├── <scope>-copy-provider.tsx
├── get-<scope>-dictionary.ts
└── resolve-request-locale.ts
```

## Database Conventions

### Tables and Columns

| Element | Convention | Example |
|---------|-----------|---------|
| Table names | Plural `snake_case` | `files`, `file_grants`, `file_views` |
| Column names | `snake_case` | `user_id`, `created_at`, `mime_type` |
| Enum types | `snake_case` | `grantee_type`, `upload_status` |
| Index names | `idx_<table>_<columns>` | `idx_files_user_created_id` |

### Migration Files

- **Timestamp prefix** format: `YYYYMMDDHHMMSS_description.{up,down}.sql`

```
20260217090948_create_files_table.up.sql
20260217090948_create_files_table.down.sql
```

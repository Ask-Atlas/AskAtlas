---
sidebar_position: 1
---

# Files API

## `GET /api/me/files`

List files owned by the authenticated user.

### Authentication

Requires a valid Clerk JWT in the `Authorization` header.

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `sort_by` | string | `updated_at` | Sort field: `updated_at`, `created_at`, `name`, `size`, `status`, `mime_type` |
| `sort_dir` | string | `desc` | Sort direction: `asc`, `desc` |
| `page_limit` | integer | `25` | Results per page (1–100) |
| `cursor` | string | — | Opaque cursor for pagination |
| `status` | string | `complete` | Filter by upload status: `pending`, `complete`, `failed` |
| `mime_type` | string | — | Filter by MIME type: `image/jpeg`, `image/png`, `image/webp`, `application/pdf` |
| `min_size` | integer | — | Minimum file size in bytes |
| `max_size` | integer | — | Maximum file size in bytes |
| `created_from` | string | — | Filter by creation date (RFC3339) |
| `created_to` | string | — | Filter by creation date (RFC3339) |
| `updated_from` | string | — | Filter by update date (RFC3339) |
| `updated_to` | string | — | Filter by update date (RFC3339) |
| `q` | string | — | Search by filename (case-insensitive substring match) |
| `scope` | string | `owned` | Access scope: `owned`, `course`, `study_guide`, `accessible` |

### Response — `200 OK`

```json
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "lecture-notes.pdf",
      "size": 2048576,
      "mime_type": "application/pdf",
      "status": "complete",
      "created_at": "2026-01-15T10:30:00Z",
      "updated_at": "2026-02-01T14:22:00Z",
      "favorited_at": "2026-01-20T08:00:00Z",
      "last_viewed_at": "2026-02-01T14:22:00Z"
    }
  ],
  "has_more": true,
  "next_cursor": "eyJpZCI6Ii4uLiJ9"
}
```

### Pagination

Uses **cursor-based (keyset) pagination**:

1. First request: omit `cursor`
2. If `has_more` is `true`, pass `next_cursor` as the `cursor` parameter in the next request
3. Repeat until `has_more` is `false`

---

## `GET /api/files/:file_id`

Get a single file by ID.

### Authentication

Requires a valid Clerk JWT in the `Authorization` header.

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `file_id` | UUID | The file ID |

### Access Control

The authenticated user can view a file if any of these are true:
- They own the file (`files.user_id = viewer_id`)
- They have a direct user grant (`file_grants.grantee_type = 'user'`)
- They belong to a course with a grant (`file_grants.grantee_type = 'course'`)
- They belong to a study guide with a grant (`file_grants.grantee_type = 'study_guide'`)
- The file is public (grant to the null UUID sentinel `00000000-0000-0000-0000-000000000000`)

### Response — `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "lecture-notes.pdf",
  "size": 2048576,
  "mime_type": "application/pdf",
  "status": "complete",
  "created_at": "2026-01-15T10:30:00Z",
  "updated_at": "2026-02-01T14:22:00Z"
}
```

---

## Shared Types

### `FileResponse`

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | File identifier |
| `name` | string | Display name |
| `size` | integer | File size in bytes |
| `mime_type` | string | `image/jpeg`, `image/png`, `image/webp`, `application/pdf` |
| `status` | string | `pending`, `complete`, `failed` |
| `created_at` | string (RFC3339) | Upload timestamp |
| `updated_at` | string (RFC3339) | Last modification timestamp |
| `favorited_at` | string (RFC3339) | When the viewer favorited this file (omitted if not favorited) |
| `last_viewed_at` | string (RFC3339) | When the viewer last viewed this file (omitted if never viewed) |

### `AppError`

All error responses use this structure:

```json
{
  "code": 400,
  "status": "Bad Request",
  "message": "Invalid query parameters",
  "details": {
    "sort_by": "must be one of: updated_at, created_at, name, size, status, mime_type",
    "min_size": "must be a non-negative integer (bytes)"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `code` | integer | HTTP status code |
| `status` | string | HTTP status text |
| `message` | string | Human-readable error message |
| `details` | object | Field-level validation errors (omitted if none) |

### Common Error Codes

| Code | Status | When |
|------|--------|------|
| 400 | Bad Request | Invalid query parameters |
| 401 | Unauthorized | Missing or invalid JWT |
| 404 | Not Found | File doesn't exist or viewer lacks access |
| 500 | Internal Server Error | Unexpected failure |

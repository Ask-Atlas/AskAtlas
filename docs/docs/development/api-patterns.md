---
sidebar_position: 1
---

# API Patterns

This guide covers the conventions and patterns used in the Go backend API.

## Endpoint Design

### Route Structure

API routes follow RESTful conventions with nested resources:

```
/api/me/files         → List files for the authenticated user
/api/files/:file_id   → Get a specific file by ID
```

### The `/me` Shorthand

`/me` is a shorthand for the authenticated user. Instead of `/api/users/:id/files`, we use `/api/me/files`. The user identity is resolved from the JWT — there's no need to pass it as a path parameter.

**When to use `/me`**: For resources scoped to the authenticated user (their files, settings, preferences).

**When to use `/resources/:id`**: For resources accessed by ID regardless of ownership (shared files, public content).

### Route Registration

Routes are registered in `cmd/api/main.go` using chi:

```go
r.Route("/api", func(r chi.Router) {
    r.Use(clerkAuth)  // All /api routes require authentication

    r.Route("/me", func(r chi.Router) {
        r.Get("/files", fileHandler.ListFiles)
    })

    r.Route("/files", func(r chi.Router) {
        r.Get("/{file_id}", fileHandler.GetFile)
    })
})
```

## Handler → Service → Repository

Every API feature follows a three-layer architecture:

```
Handler (HTTP)  →  Service (Business Logic)  →  Repository (Database)
```

### Handler Layer (`internal/handlers/`)

Handles HTTP concerns only:
- Parse and validate request parameters
- Call the service
- Map service errors to HTTP responses
- Encode JSON responses

```go
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
    params, appErr := parseListFilesParams(r)
    if appErr != nil {
        apperrors.RespondWithError(w, appErr)
        return
    }

    files, nextCursor, err := h.service.ListFiles(r.Context(), *params)
    if err != nil {
        appErr := apperrors.ToHTTPError(err)
        apperrors.RespondWithError(w, appErr)
        return
    }

    respondJSON(w, http.StatusOK, response)
}
```

### Service Layer (`internal/<domain>/service.go`)

Contains business logic and data transformation:
- Validates business rules
- Orchestrates repository calls
- Converts between domain models and database types

The service depends on a **Repository interface**, not a concrete implementation.

### Repository Layer (`internal/<domain>/sqlc_repository.go`)

Thin adapter between the service and sqlc-generated code:
- Calls sqlc-generated queries
- Maps database rows to domain models

## Interface-Driven Design

Each layer defines the interface it depends on in the **consuming** package:

```go
// In handlers/ — defines what it needs from the service
type FileService interface {
    GetFile(ctx context.Context, params files.GetFileParams) (files.File, error)
    ListFiles(ctx context.Context, params files.ListFilesParams) ([]files.File, *string, error)
}
```

This allows mock generation with [mockery](https://github.com/vektra/mockery) for unit testing:

```bash
make mockery
```

Mocks are generated into `internal/<domain>/mocks/`.

## Error Handling (`pkg/apperrors`)

All API errors use the structured `AppError` type:

```go
type AppError struct {
    Code    int               `json:"code"`
    Status  string            `json:"status"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}
```

### Error Factories

| Function | HTTP Code | When |
|----------|-----------|------|
| `NewBadRequest(msg, details)` | 400 | Invalid input with field-level errors |
| `NewUnauthorized()` | 401 | Missing or invalid auth |
| `NewForbidden()` | 403 | Valid auth but insufficient permissions |
| `NewNotFound(msg)` | 404 | Resource not found |
| `NewInternalError()` | 500 | Unexpected failures |

### Error Flow

1. **Service** returns a sentinel error (e.g. `apperrors.ErrNotFound`)
2. **Handler** calls `apperrors.ToHTTPError(err)` to map it to an `AppError`
3. **Handler** calls `apperrors.RespondWithError(w, appErr)` to write the JSON response
4. **5xx errors** are logged with `slog.Error` for observability

## sqlc Workflow

We use [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

1. Write SQL queries in `db/queries/<domain>.sql` using sqlc annotations
2. Run `sqlc generate` (configured in `sqlc.yml`)
3. Generated code appears in `internal/db/`
4. Repository adapters in `internal/<domain>/sqlc_repository.go` call the generated functions

## Adding a New Endpoint

1. **Write the SQL query** in `db/queries/<domain>.sql`
2. **Run `sqlc generate`** to create the Go database functions
3. **Add Repository method** in `internal/<domain>/sqlc_repository.go`
4. **Add Repository interface method** in `internal/<domain>/service.go`
5. **Write Service method** with business logic
6. **Write Handler method** with HTTP parsing and response
7. **Register the route** in `cmd/api/main.go`
8. **Generate mocks** with `make mockery`
9. **Write tests** for handler and service

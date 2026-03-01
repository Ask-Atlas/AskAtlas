---
sidebar_position: 6
---

# Error Handling Strategy

The API uses a structured error handling approach with the `pkg/apperrors` package.

## Architecture

```
Repository → returns sentinel errors or wrapped errors
     ↓
Service → wraps with context via fmt.Errorf
     ↓
Handler → maps to AppError via ToHTTPError, responds with JSON
```

## AppError

All API error responses use the `AppError` struct:

```go
type AppError struct {
    Code    int               `json:"code"`
    Status  string            `json:"status"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}
```

The `Details` field provides **field-level validation errors**, making it easy for frontend consumers to display errors next to specific form fields.

## Sentinel Errors

Define domain-level sentinel errors for expected failure cases:

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrConflict     = errors.New("already exists")
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
    ErrInvalidInput = errors.New("invalid input")
)
```

## Error Flow

### 1. Repository Layer

Wrap errors with the function name and use sentinel errors for expected cases:

```go
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return File{}, fmt.Errorf("GetFileIfViewable: %w", apperrors.ErrNotFound)
    }
    return File{}, fmt.Errorf("GetFileIfViewable: %w", err)
}
```

### 2. Service Layer

Wrap errors with additional context. Don't log — let the handler handle it:

```go
if err != nil {
    return nil, nil, fmt.Errorf("ListFiles: %w", err)
}
```

### 3. Handler Layer

Map errors to HTTP responses using `ToHTTPError`:

```go
file, err := h.service.GetFile(ctx, params)
if err != nil {
    slog.Error("GetFile failed", "error", err)
    appErr := apperrors.ToHTTPError(err)
    apperrors.RespondWithError(w, appErr)
    return
}
```

### 4. Validation Errors

For request validation, return a `400 Bad Request` with field-level details:

```go
details := map[string]string{}
if invalidSort {
    details["sort_by"] = "must be one of: updated_at, created_at, name, size"
}
if invalidLimit {
    details["page_limit"] = "must be an integer between 1 and 100"
}
if len(details) > 0 {
    return nil, apperrors.NewBadRequest("Invalid query parameters", details)
}
```

## Error Mapping Table

`ToHTTPError` maps sentinel errors to HTTP status codes:

| Sentinel | HTTP Code | Status |
|----------|-----------|--------|
| `ErrNotFound` | 404 | Not Found |
| `ErrConflict` | 409 | Conflict |
| `ErrInvalidInput` | 400 | Bad Request |
| `ErrUnauthorized` | 401 | Unauthorized |
| `ErrForbidden` | 403 | Forbidden |
| Unknown/unexpected | 500 | Internal Server Error |

## Rules

1. **Never expose internal error messages to the client** — `NewInternalError()` returns a generic "Something went wrong" message
2. **Always wrap errors** with the function name using `fmt.Errorf("FunctionName: %w", err)`
3. **Log at the handler level** — the handler is the only layer that logs errors (see [Logging Patterns](./logging-patterns))
4. **Use `Details` for validation** — provide per-field error messages so the frontend can show them inline
5. **Use sentinels for expected cases** — `ErrNotFound`, `ErrConflict`, etc. enable proper HTTP status mapping

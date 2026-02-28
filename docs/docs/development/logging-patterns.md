---
sidebar_position: 5
---

# Logging Patterns

The API uses Go's standard [`log/slog`](https://pkg.go.dev/log/slog) package for structured logging.

## Golden Rule: Log Once Per Error

Every error should be logged **exactly once** — at the HTTP boundary (handler or middleware). Inner layers (service, repository) **wrap and return** errors but do not log them.

### Why?

- **No duplicate log lines** — If every layer logs the same error, you get N lines for one failure. Noisy and confusing.
- **Full context in one place** — The handler has the most context (HTTP method, path, user ID) to produce a useful log line.
- **Clean separation** — Inner layers focus on building a descriptive error chain. The boundary focuses on observability.

## Error Wrapping: Build the Trace

Each layer wraps the error with its function name using `fmt.Errorf`. This builds a human-readable trace without needing stack traces:

```go
// Repository
return File{}, fmt.Errorf("GetFileIfViewable: %w", err)

// Service
return nil, fmt.Errorf("ListFiles: %w", err)
```

The handler receives a chained error like:

```
ListFiles: dispatchListQuery: ListOwnedFilesUpdatedDesc: connection refused
```

This tells you exactly what happened and where — without grep-searching log lines.

### Wrapping Format

```
fmt.Errorf("FunctionName: %w", err)
```

- Use the function or method name — not a freeform description
- Always use `%w` to preserve the original error for `errors.Is()` / `errors.As()` checks

## Log Levels

| Level | Purpose | Who Uses It |
|-------|---------|-------------|
| `slog.Debug` | Operation tracing (what query ran, what ID was used) | Repository |
| `slog.Info` | Meaningful lifecycle events | `main.go`, webhook receipt |
| `slog.Warn` | Recoverable issues, degraded state | Middleware (auth failures) |
| `slog.Error` | Failures that affect the response | Handler only |

### When NOT to Use Each Level

| Level | Avoid |
|-------|-------|
| `Debug` | Don't use for routine success — `slog.Debug("file found")` is noise |
| `Info` | Don't use for every database call — that's `Debug`. Reserve for events worth tracking in production. |
| `Warn` | Don't use for expected business logic (e.g. "file not found" is a 404, not a warning) |
| `Error` | Don't use outside handlers/middleware — inner layers wrap and return |

## Handler Logging Pattern

Only log **5xx errors**. A 404 is not an error — it's a normal response.

```go
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
    file, err := h.service.GetFile(r.Context(), params)
    if err != nil {
        appErr := apperrors.ToHTTPError(err)
        if appErr.Code >= 500 {
            slog.Error("GetFile failed", "error", err)
        }
        apperrors.RespondWithError(w, appErr)
        return
    }
    respondJSON(w, http.StatusOK, files.ToFileResponse(file))
}
```

**Key**: The `if appErr.Code >= 500` guard ensures only unexpected failures produce log lines. Expected errors (404, 400) are returned to the client without logging noise.

## Structured Fields

Always use key-value pairs — never interpolate into the message string:

```go
// ✅ Good — structured, queryable, parseable
slog.Error("ClerkAuth: failed to resolve user ID",
    "clerk_id", clerkUserID,
    "error", err,
)

// ❌ Bad — unstructured, impossible to filter
slog.Error(fmt.Sprintf("failed for user %s: %v", clerkUserID, err))
```

### Standard Field Names

Use consistent keys across the codebase:

| Key | Type | When |
|-----|------|------|
| `"error"` | `error` | Always include the error value |
| `"clerk_id"` | `string` | Clerk user identity |
| `"file_id"` | `uuid` | File being operated on |
| `"owner_id"` | `uuid` | Resource owner |
| `"type"` | `string` | Event type (webhooks) |

## Context-Aware Logging

For request-scoped logs, use `slog.InfoContext` / `slog.ErrorContext` to propagate request context:

```go
slog.InfoContext(ctx, "received webhook event", "type", eventType)
slog.ErrorContext(ctx, "failed to handle webhook event", "type", eventType, "error", err)
```

This enables future middleware to inject request IDs or trace IDs into all log output automatically.

## Summary

```
Repository  →  slog.Debug for tracing, fmt.Errorf to wrap errors
Service     →  fmt.Errorf to wrap errors, no logging
Handler     →  slog.Error for 5xx only, respond with AppError
Middleware  →  slog.Warn for auth failures, slog.Error for infra failures
main.go     →  slog.Info for startup, slog.Error for fatal config
```

---
sidebar_position: 7
---

# Testing Strategy

## Overview

Every new feature should include both **unit tests** and **E2E tests**. New endpoints must have corresponding E2E tests covering success and error contracts.

## API Unit Tests

Unit tests live alongside the code they test with the `_test.go` suffix.

### What Gets Tested

| Component | What to Test | Example File |
|-----------|-------------|-------------|
| Service | Business logic, scope guards, data transformation | `internal/files/service_test.go` |
| Handler | HTTP parsing, response codes, error mapping | `internal/handlers/files_test.go` |
| Params | Query parameter validation, edge cases | `internal/files/params_test.go` |
| Mappers | Domain ↔ DB model conversion | `internal/clerk/mapper_test.go`, `internal/user/mapper_test.go` |
| Auth context | Context helpers | `pkg/authctx/authctx_test.go` |

### Running Unit Tests

```bash
cd api
make test
```

### Using Mocks

Interfaces are mocked with [mockery](https://github.com/vektra/mockery). Generate mocks after adding or changing an interface:

```bash
make mockery
```

Mocks are generated into `internal/<domain>/mocks/`. Use them in tests to isolate the layer under test:

```go
func TestGetFile_NotFound(t *testing.T) {
    mockService := mocks.NewFileService(t)
    mockService.On("GetFile", mock.Anything, mock.Anything).
        Return(files.File{}, apperrors.ErrNotFound)
    
    handler := NewFileHandler(mockService)
    // ... test the handler returns 404
}
```

### Test Conventions

- Use `testify/assert` and `testify/require`
- Group related tests with `t.Run()` subtests
- Name tests descriptively: `TestListFiles_InvalidSortBy_Returns400`

## API E2E Tests

E2E tests live in `api/e2e/tests/` and use [Playwright](https://playwright.dev/) to make real HTTP requests against a running API.

### What Gets Tested

| File | Coverage |
|------|----------|
| `contracts.spec.ts` | Response shape validation (DTO fields, types) |
| `files.spec.ts` | Filter/sort/pagination behavior, error codes |

### What Every New Endpoint Needs

1. **Success contracts** — Verify the response shape matches the DTO
2. **Error contracts** — Verify 400, 401, 404 responses have the correct `AppError` shape
3. **Validation coverage** — Test all query parameter validation rules
4. **Pagination** — If the endpoint is paginated, test cursor traversal

### Running E2E Tests

```bash
cd api
make e2e E2E_TOKEN=<your-jwt-token>
```

The `E2E_TOKEN` is a valid Clerk JWT for an authenticated test user. You can obtain one from the browser's network tab or the Clerk dashboard.

## Frontend Tests

Frontend E2E tests use Playwright and live in `web/e2e/`.

### Running Frontend Tests

```bash
cd web
make test-e2e          # against dev
make test-e2e-staging  # against staging
make test-e2e-prod     # against production
make e2e-report        # view the HTML report
make test-e2e-codegen  # open Playwright codegen for recording tests
```

## When to Write Tests

| Scenario | Unit Tests | E2E Tests |
|----------|-----------|-----------|
| New API endpoint | ✅ Handler + Service + Params | ✅ Success + error contracts |
| New service logic | ✅ Service | — |
| New query parameter | ✅ Params validation | ✅ E2E validation |
| Bug fix | ✅ Regression test | ✅ If it's an API-level bug |
| Refactor | ✅ Verify existing tests pass | — |
| New frontend page | — | ✅ Navigation + key interactions |

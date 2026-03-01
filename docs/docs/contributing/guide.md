---
sidebar_position: 1
---

# Contributing Guide

## Commit Convention

We use **semantic commits** to keep the git history readable and enable automated tooling.

### Format

```
<type>: <description>
```

### Types

| Type | Purpose | Example |
|------|---------|---------|
| `feat` | New feature | `feat: add cursor-based pagination to file listing` |
| `fix` | Bug fix | `fix: resolve race condition in webhook user creation` |
| `docs` | Documentation only | `docs: add architecture overview to Docusaurus` |
| `chore` | Maintenance, deps, config | `chore: bump Go to 1.24.1` |
| `refactor` | Code change with no feature or fix | `refactor: extract file permission checks into service layer` |
| `test` | Adding or updating tests | `test: add e2e tests for file filter validation` |
| `ci` | CI/CD pipeline changes | `ci: add rollback workflow for API deployments` |

### Rules

- Use **lowercase** for the type prefix
- Use **imperative mood** for the description ("add", not "added" or "adds")
- Keep the first line under 72 characters
- Reference issue numbers in the body if applicable

## Pull Requests

### Requirements to Merge

Before a Pull Request can be merged into the default branch, it must meet the following criteria:

1. **One Approval:** At least 1 review approval is required from another team member.
2. **Passing Checks:** All automated CI/CD status checks (tests, linters, deployments) must pass.
3. **Up to Date:** The PR branch must be up to date with the default branch.

### Squash Merges

All PRs must be **squash-merged** to keep the main branch history clean. The squash commit message should follow the semantic commit format.

### PR Template

When you open a PR, the template is auto-populated. Fill in each section:

1. **Description** — What the PR does and why. Reference the issue with `Closes #<number>`.
2. **Type of change** — Check the appropriate box.
3. **Changes** — Group changes by area (e.g., Auth, Files domain, Infrastructure). Be specific.
4. **How has this been tested?** — List unit, integration, and/or E2E tests.
5. **Checklist** — Confirm self-review, tests, no new warnings, and tidy modules.

### Example PR

```markdown
## Description

Introduces the files domain with authenticated read access.
Wires together Clerk JWT auth, a new `files` package, and
supporting infrastructure for `GET /api/me/files` and
`GET /api/files/:file_id`.

Closes #46

## Type of change

- [x] New feature (non-breaking)

## Changes

**Auth**
- `ClerkAuth` middleware — validates Clerk JWTs, resolves user ID
- `authctx` package — typed context helpers

**Files domain** (`internal/files`)
- Domain model, repository, service, DTOs, HTTP handler
- `GET /api/me/files` — cursor-based pagination, sorts, filters
- `GET /api/files/:file_id` — fetch single viewable file

## How has this been tested?

- Unit — files service, handler, param parsing
- E2E — Playwright tests for success and error contracts

## Checklist

- [x] Self-reviewed code
- [x] Added/updated tests
- [x] No new warnings introduced
- [x] `go mod tidy` run and verified
```

## Code Review

- **Review before requesting** — Self-review your diff before marking ready
- **Keep PRs focused** — One feature or fix per PR when possible
- **Respond to feedback** — Address all comments before merging

# Contributing to AskAtlas

Thanks for contributing! This guide covers the essentials — see the [full contributing docs](https://ask-atlas.github.io/AskAtlas/docs/contributing/guide) for details.

## Commit Convention

We use **semantic commits**. Format: `<type>: <description>`

| Type | Purpose |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Maintenance, deps, config |
| `refactor` | Code change with no feature or fix |
| `test` | Adding or updating tests |
| `ci` | CI/CD pipeline changes |

**Examples:**
```
feat: add cursor-based pagination to file listing
fix: resolve race condition in webhook user creation
docs: add architecture overview to Docusaurus
refactor: extract file permission checks into service layer
test: add e2e tests for file filter validation
ci: add rollback workflow for API deployments
chore: bump Go to 1.24.1
```

## Pull Requests

- All PRs must be **squash-merged** to keep history clean
- All PRs require at least **1 approval** from a reviewer
- All CI status checks (tests, linters, deployments) must pass before merging
- Use the [PR template](.github/pull_request_template.md) — it's auto-populated when you open a PR
- Self-review your code before requesting review

## Development

See the [Getting Started guide](https://ask-atlas.github.io/AskAtlas/docs/getting-started/prerequisites) for environment setup.

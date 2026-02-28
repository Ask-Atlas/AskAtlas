# AskAtlas

All-in-one student workspace for planning, practice, and class collaboration.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | [Next.js](https://nextjs.org) (App Router) |
| Backend | [Go](https://go.dev) + [chi](https://github.com/go-chi/chi) |
| Database | [PostgreSQL](https://www.postgresql.org) |
| Auth | [Clerk](https://clerk.com) |
| Secrets | [Infisical](https://infisical.com) |
| CI/CD | GitHub Actions → Docker → [Digital Ocean](https://digitalocean.com) |

## Repository Layout

```
AskAtlas/
├── api/            # Go backend API
├── web/            # Next.js frontend
├── migrations/     # PostgreSQL migration scripts
├── docs/           # Docusaurus documentation site
├── scripts/        # Deployment and rollback scripts
└── .github/        # CI/CD workflows and PR template
```

## Getting Started

1. Install [prerequisites](https://ask-atlas.github.io/AskAtlas/docs/getting-started/prerequisites)
2. Set up each service:
   - **API**: see [API Setup](https://ask-atlas.github.io/AskAtlas/docs/getting-started/api-setup)
   - **Web**: see [Web Setup](https://ask-atlas.github.io/AskAtlas/docs/getting-started/web-setup)
   - **Database**: see [Database Setup](https://ask-atlas.github.io/AskAtlas/docs/getting-started/database-setup)

## Documentation

Full documentation is available at **[ask-atlas.github.io/AskAtlas](https://ask-atlas.github.io/AskAtlas/)**, covering architecture, infrastructure, development patterns, API reference, and contributing guidelines.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for commit conventions, PR guidelines, and development workflow.

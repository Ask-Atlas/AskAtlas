---
slug: sprint-1-recap
title: "Sprint 1: From Zero to Foundation"
authors: [askatlas]
tags: [sprint]
---

We just wrapped Sprint 1 and wanted to share what we built, what we learned, and where we're headed.

<!-- truncate -->

## What We Built

Starting from an empty repo, we shipped:

- **Full authentication pipeline** — Clerk integration on both frontend (Next.js middleware) and backend (Go JWT validation + webhook sync)
- **Database architecture** — PostgreSQL with a comprehensive schema for users and files, including grants, views, favorites, and soft deletes
- **CI/CD pipelines** — Automated checks for both API and web, Docker-based deployments to Digital Ocean, and rollback capability
- **Core UI** — Landing page, authenticated dashboard, practice page designs, and responsive sidebar navigation

All 19 planned issues were completed.

## Architecture Highlights

### Split-Stack Monorepo

We chose a Go API + Next.js frontend in a monorepo. Go gives us strong typing and performance for the API, while Next.js with the App Router lets us leverage server components and server-side rendering for the frontend.

### Interface-Driven Go Patterns

The API follows a clean **Handler → Service → Repository** architecture. Every layer depends on interfaces, making unit testing straightforward with generated mocks.

### Keyset Pagination

For file listings, we implemented cursor-based (keyset) pagination instead of offset pagination. This is more efficient at scale — instead of `OFFSET N` (which scans and discards N rows), we use `WHERE (sort_col, id) > (cursor_value, cursor_id)` which seeks directly via the index.

## What We Learned

This sprint had a significant learning curve. The tech stack — Go, Next.js App Router, PostgreSQL, Docker, Clerk, Infisical — was new for part of the team. We spent considerable time on knowledge transfer and onboarding.

Key takeaways:
- **Document early** — We didn't document architecture decisions during the sprint, which made it harder for team members to ramp up
- **Pair programming helps** — Working through unfamiliar code together was more effective than reading docs alone
- **Infrastructure first pays off** — Investing in CI/CD and deployment automation upfront has already saved us time

## What's Next

Sprint 2 focuses on building out the core CRUD operations now that the foundation is solid:
- File upload/download pipeline via Garage (S3-compatible storage)
- Course and study guide CRUD
- Library UI/UX
- Dashboard implementation with real data

For the full sprint report with issue details, see the [Sprint 1 Review](/docs/sprint-reviews/sprint-1).

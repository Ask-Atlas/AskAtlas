---
slug: sprint-2-recap
title: "Sprint 2: Building the Core Experience"
authors: [askatlas]
tags: [sprint]
---

Sprint 2 focused on turning our foundation into something users can actually interact with — file management, study guide views, and a real dashboard.

<!-- truncate -->

## What We Built

Building on Sprint 1's infrastructure, we shipped:

- **File deletion with async workers** — DELETE endpoint queues work to a background worker instead of blocking the request. Files are soft-deleted for recoverability.
- **File renaming** — PATCH endpoint for updating file metadata, starting with the ability to rename files.
- **Auto-generated API docs** — Migrated from manually maintained API reference files to OpenAPI-generated documentation that stays in sync with the codebase.
- **Dashboard, Study Guide View, and Library UIs** — Three major frontend pages progressed from design to implementation, though still being connected to live data.

## Architecture Highlights

### Async File Deletion

Instead of deleting files synchronously in the API handler, we implemented a worker-based flow. The API marks the file for deletion and enqueues a job. A background worker picks it up and handles the actual cleanup. This keeps API response times predictable and sets a pattern we can reuse for other heavy operations like file processing.

### Soft Deletes

Files are now soft-deleted rather than permanently removed. The repository layer filters out deleted records by default, but they remain in the database for recovery. This is a safer pattern for user-facing delete operations.

### OpenAPI as Source of Truth

We migrated from hand-written API documentation to auto-generated docs from our OpenAPI spec. The Docusaurus site now pulls API reference content directly from the spec, so docs stay accurate as endpoints change.

## What's Next

Sprint 3 will focus on closing the loop on features that are in progress and building out the content pipeline:

- Merge open PRs for dashboard, study guide view, library, and file renaming
- File upload pipeline via Garage (S3-compatible storage)
- Course and study guide CRUD operations
- Connecting frontend pages to live API data

For the full sprint report with issue details, see the [Sprint 2 Review](/docs/sprint-reviews/sprint-2).

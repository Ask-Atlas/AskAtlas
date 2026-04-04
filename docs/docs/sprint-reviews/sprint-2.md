---
sidebar_position: 2
---

# Sprint 2 Report (03/02/26 – 04/04/26)

[**Sprint demo video**](TODO)

## What's New (User Facing)

- **File Deletion** — Users can now delete files they own, with soft-delete support so files can be recovered if needed
- **File Renaming** — Users can rename their uploaded files via the updated file management interface
- **Study Guide View** — New dedicated page for viewing study guides with feature components for an improved reading experience
- **Dashboard Redesign** — Redesigned authenticated dashboard with statistics cards (total study guides, completed quizzes, study streak), activity timeline, and recent study guide viewer
- **Resource Library UI** — New library interface for browsing and discovering shared resources

## Work Summary (Developer Facing)

This sprint focused on building out core CRUD operations and user-facing pages on top of the foundation established in Sprint 1.

On the backend, we implemented the DELETE endpoint for files with an asynchronous worker flow — deletion requests are queued and processed by a background worker, keeping the API responsive. The file repository was updated with soft-delete support so records are marked as deleted rather than permanently removed. We also added a PATCH endpoint for updating file metadata (renaming). A major infrastructure improvement was migrating the API documentation from manually maintained files to auto-generated content using OpenAPI specs, which keeps docs in sync with the actual API surface.

On the frontend, significant progress was made on three key pages: the authenticated dashboard with real UI components (stat cards, activity timeline, recent study guides), the study guide view page, and the resource library. These pages are still being connected to live data but the UI and interaction patterns are in place.

CI/CD was improved by scoping temporary script directories to individual workflow runs, preventing conflicts in parallel builds. Backend Go source code documentation was automated, reducing manual upkeep.

## Completed Issues

| Issue | Description |
|-------|-------------|
| [#48](https://github.com/Ask-Atlas/AskAtlas/issues/48) | API - DELETE /api/files/:id |
| [PR #52](https://github.com/Ask-Atlas/AskAtlas/pull/52) | Migrate API service to OpenAPI |
| [PR #54](https://github.com/Ask-Atlas/AskAtlas/pull/54) | CI: scope temporary script directories to prevent conflicts |
| [PR #56](https://github.com/Ask-Atlas/AskAtlas/pull/56) | Automate backend Go API documentation |

## In Progress

| Issue | PR | Description |
|-------|-----|-------------|
| [#6](https://github.com/Ask-Atlas/AskAtlas/issues/6) / [#7](https://github.com/Ask-Atlas/AskAtlas/issues/7) | [#57](https://github.com/Ask-Atlas/AskAtlas/pull/57) | Dashboard UI/UX design and implementation |
| [#15](https://github.com/Ask-Atlas/AskAtlas/issues/15) | [#62](https://github.com/Ask-Atlas/AskAtlas/pull/62) | Study guide view page and feature components |
| [#61](https://github.com/Ask-Atlas/AskAtlas/issues/61) | [#63](https://github.com/Ask-Atlas/AskAtlas/pull/63) | API - PATCH /files/{file_id} (update file metadata) |
| [#19](https://github.com/Ask-Atlas/AskAtlas/issues/19) | [#64](https://github.com/Ask-Atlas/AskAtlas/pull/64) | Library UI/UX design |

## Planned for Future Sprints

| Issue | Description |
|-------|-------------|
| [#60](https://github.com/Ask-Atlas/AskAtlas/issues/60) | API - POST /api/files (create file reference + presigned upload) |
| [#47](https://github.com/Ask-Atlas/AskAtlas/issues/47) | API - POST/DELETE /api/files/{file_id}/grants (manage file grants) |
| [#45](https://github.com/Ask-Atlas/AskAtlas/issues/45) | Migration - Course & Study Guide Join Tables |
| [#17](https://github.com/Ask-Atlas/AskAtlas/issues/17) | CRUD for library resources |
| [#14](https://github.com/Ask-Atlas/AskAtlas/issues/14) | CRUD for Quizzes |
| [#13](https://github.com/Ask-Atlas/AskAtlas/issues/13) | Scrape Wazzu for Courses |
| [#12](https://github.com/Ask-Atlas/AskAtlas/issues/12) | CRUD for study guides |
| [#11](https://github.com/Ask-Atlas/AskAtlas/issues/11) | Scrape Wazzu for courses |
| [#10](https://github.com/Ask-Atlas/AskAtlas/issues/10) | CRUD for Courses |
| [#9](https://github.com/Ask-Atlas/AskAtlas/issues/9) | CRUD Operations API Users |

## Retrospective

### What went well
- Established async worker pattern for file deletion — sets a reusable pattern for future background processing
- Automated API documentation keeps docs in sync without manual effort
- Multiple frontend pages progressed in parallel across team members
- OpenAPI migration gives us a single source of truth for API contracts

### What we'd like to improve
- Several PRs remain open at sprint end — need tighter scoping or earlier merges
- Frontend pages need to be connected to real backend data sooner in the development cycle
- More issues should have been formally tracked — some work (OpenAPI migration, CI fixes) was done as PRs without corresponding GitHub issues

### Changes for next sprint
- Set mid-sprint checkpoints to identify PRs at risk of not merging by sprint end
- Prioritize connecting frontend pages to live API endpoints
- Create GitHub issues before starting work, even for infrastructure improvements
- Begin implementing file upload pipeline to complete the file management feature set

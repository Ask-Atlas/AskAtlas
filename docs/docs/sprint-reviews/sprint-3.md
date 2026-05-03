---
sidebar_position: 3
---

# Sprint 3 Report (04/05/26 – 05/02/26)

[**Sprint demo video**](https://youtu.be/0BMBwl0C34E)

## What's New (User Facing)

- **AI-powered study guide editing** — Selection toolbar, prompt input, and per-hunk diff overlay for accepting or rejecting changes
- **Preset AI edit actions** — TL;DR, reorganize, and rewrite-for-audience
- **Study guide visibility controls** — Publish/draft states and a share UI for managing grants
- **Study guide article renderer** — Markdown, images, and internal links
- **Study guide creation and editing** — Pages fully wired up, including a shared form with write/preview tab and paste-URL rewriting
- **Study guide cards** — List and compact variants
- **File downloads** — Via presigned GET URLs
- **File row menu** — Inline rename and delete
- **Courses** — Catalog, detail, and section membership pages wired up with join/leave functionality
- **Practice page** — Wired up to real quiz session APIs
- **Shared UI primitives** — Confirmation dialog, toast provider, favorite button with optimistic updates, skeleton loaders, empty state, and error boundary pages

## Work Summary (Developer Facing)

Sprint 3 was the team's most technically ambitious sprint to date, centered on building out AskAtlas's AI layer and completing a large volume of frontend component and page wire-up work. David delivered the full backend AI pipeline, Anthropic client and SSE handler, streaming AI edit endpoint with audit table, AI rate limiting and cost ledger, file text extraction via QStash, and a chunk-and-embed worker using OpenAI's text-embedding-3-small model. On the frontend, David also completed the diff overlay UI for per-hunk accept/reject of AI edits, preset edit actions, study guide visibility and share UI, and numerous shared components and wired-up pages. Nathaniel contributed the study guide card component and the UI/UX designs for the study guide view and practice page. Luca handled the dashboard and library UI/UX designs and the initial dashboard implementation. The pgvector schema for study guide file chunks (ASK-219) is currently still in progress and will carry into the next sprint. The Todo queue remains large, with the bulk of remaining frontend wire-up and the AI feature backlog (Q&A panel, AI quiz generation, inline citations) slated for Sprint 4.

## Unfinished Work

| Issue | Reason |
|-------|--------|
| ASK-219: DB - pgvector + study_guide_file_chunks schema | Deprioritized as the sprint progressed in favor of completing the higher-priority AI pipeline and frontend work. The pgvector schema would serve as a foundation for retrieval and citation features that remain in the backlog and are outside the scope of the final deliverable. |

## Completed Issues/User Stories

| Issue | Description |
|-------|-------------|
| ASK-215 | Backend - AI edit endpoint (stream) + edit audit table |
| ASK-221 | Backend - chunk + embed worker (OpenAI text-embedding-3-small) |
| ASK-217 | Frontend - diff overlay + per-hunk accept/reject for AI edits |
| ASK-213 | Backend - AI plumbing: Anthropic client + SSE handler |
| ASK-220 | Backend - file text extraction worker (QStash) |
| ASK-214 | Backend - AI rate limit + cost ledger |
| ASK-216 | Frontend - selection toolbar + AI prompt input |
| ASK-218 | Frontend - preset edit actions (TL;DR, reorganize, rewrite-for-audience) |
| ASK-211 | Backend - study-guide visibility enforcement + grants CRUD |
| ASK-212 | Frontend - study-guide visibility + share UI |
| ASK-207 | Backend + Frontend - study-guide visibility / publish / grants |
| ASK-208 | Backend - batch POST `/api/refs/resolve` for inline entity cards |
| ASK-209 | Frontend - custom-element remark plugin + entity ref cards + callout block |
| ASK-210 | Frontend - StudyGuideForm paste-URL rewriter + write/preview tab |
| ASK-206 | Frontend - study-guides/article-renderer (markdown + images + internal links) |
| ASK-205 | Backend - file download: presigned GET + `download_url` on FileResponse |
| ASK-195 | Frontend - study-guides/edit-page-wire-up |
| ASK-191 | Frontend - study-guides/new-page-wire-up |
| ASK-190 | Frontend - study-guides/view-page-wire-up (supersedes ASK-135) |
| ASK-170 | Frontend - study-guides/study-guide-form (shared create + edit) |
| ASK-169 | Frontend - study-guides/study-guide-card (list + compact) |
| ASK-165 | Frontend - files/file-row-menu (inline rename + delete) |
| ASK-164 | Frontend - files/file-card (list + grid variants) |
| ASK-197 | Frontend - courses/detail-page-wire-up |
| ASK-193 | Frontend - courses/catalog-page-wire-up |
| ASK-182 | Frontend - courses/section-membership-button (join / leave) |
| ASK-181 | Frontend - courses/course-search-bar |
| ASK-180 | Frontend - courses/course-card (row + tile variants) |
| ASK-192 | Frontend - practice/page-wire-up (supersedes ASK-126) |
| ASK-175 | Frontend - quizzes/quiz-card |
| ASK-163 | Frontend - shared/confirmation-dialog primitive |
| ASK-168 | Frontend - shared/favorite-button (useOptimistic star toggle) |
| ASK-162 | Frontend - shared/skeleton-primitives (list + grid) |
| ASK-161 | Frontend - shared/empty-state component |
| ASK-160 | Frontend - shared/error-boundary-pages: `error.tsx` + `not-found.tsx` conventions |
| ASK-159 | Frontend - shared/toast-provider: sonner + typed toast helpers |
| ASK-158 | Frontend - shared/shadcn-install-missing: install 9 missing primitives + react-hook-form |

## Incomplete Issues/User Stories

| Issue | Reason |
|-------|--------|
| ASK-219: DB - pgvector + study_guide_file_chunks schema | Still in progress; will be completed at the start of Sprint 4 as it unblocks retrieval and citation features. |

## Code Files for Review

Please review the following code files, which were actively developed during this sprint, for quality:

| File | Description |
|------|-------------|
| `ai/handler.go` | SSE streaming handler for AI edit endpoint |
| `ai/ratelimit.go` | AI rate limiting and cost ledger logic |
| `workers/embed.go` | Chunk and embed worker using OpenAI text-embedding-3-small |
| `workers/extract.go` | File text extraction worker via QStash |
| `study-guides/article-renderer.tsx` | Markdown article renderer with image and internal link support |
| `study-guides/study-guide-form.tsx` | Shared create/edit form with write/preview tab |
| `shared/diff-overlay.tsx` | Per-hunk accept/reject UI for AI edits |

## Retrospective

### What went well

- The AI pipeline came together faster than expected — having a clear technical plan going into the sprint meant David could move through the backend work without significant blockers
- Frontend and backend development continued to stay largely in sync, with shared components and page wire-ups landing at the right time to support the AI features built on top of them
- The team had a strong sense of what the final product needed to look like this sprint, which made prioritization decisions easier and kept everyone focused
- Closing out the core user journeys end to end was a meaningful milestone — the app finally feels like a real product

### What we'd like to improve

- The Todo queue grew larger than ideal over the course of the project, and some frontend wire-up work that would have added polish did not make it into the final sprint
- PR reviews occasionally lagged, especially during the heavier backend weeks when everyone was heads-down on their own work
- Some backlog features like the Q&A panel and AI quiz generation were ambitious ideas that we never found the right sprint to tackle

### Changes for next sprint

This was the final sprint of the project. If development were to continue, the immediate priorities would be completing the pgvector schema, wiring the retrieval pipeline to the frontend, and tackling the AI quiz generation and Q&A panel features that remained in the backlog throughout the project lifecycle.

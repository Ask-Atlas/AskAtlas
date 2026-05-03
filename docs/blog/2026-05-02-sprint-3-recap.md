---
slug: sprint-3-recap
title: "Sprint 3: Shipping the AI Layer"
authors: [askatlas]
tags: [sprint]
---

Sprint 3 was the team's most technically ambitious sprint, centered on building out AskAtlas's AI layer end to end and closing out the remaining frontend wire-up so the app finally behaves like a real product.

<!-- truncate -->

## What We Built

The AI features came together this sprint, alongside a large batch of frontend pages and shared primitives:

- **AI study guide editing** — Selection toolbar, prompt input, and a per-hunk diff overlay for accepting or rejecting changes, plus presets for TL;DR, reorganize, and rewrite-for-audience.
- **AI backend pipeline** — Anthropic client and SSE handler, streaming edit endpoint with audit table, rate limiting and a cost ledger, file text extraction via QStash, and a chunk + embed worker on OpenAI's `text-embedding-3-small`.
- **Study guides end to end** — Article renderer (markdown, images, internal links), shared create/edit form with write/preview and paste-URL rewriting, visibility controls with publish/draft and a share UI for grants, and list + compact card variants.
- **Files** — Presigned GET downloads and an inline rename/delete row menu.
- **Courses** — Catalog, detail, and section membership pages wired up with join/leave.
- **Practice** — Page wired up to the real quiz session APIs.
- **Shared primitives** — Confirmation dialog, toast provider, optimistic favorite button, skeleton loaders, empty state, and error boundary pages.

## How We Worked

David delivered the full backend AI pipeline and the bulk of the frontend wire-up, including the diff overlay, preset edit actions, visibility/share UI, and the shared primitives. Nathaniel contributed the study guide card and the UI/UX designs for the study guide view and practice page. Luca handled the dashboard and library UI/UX designs and the initial dashboard implementation.

A clear technical plan going into the sprint meant the AI work moved faster than expected. Frontend and backend stayed largely in sync — shared components and page wire-ups landed at the right time to support the AI features built on top of them.

## What Didn't Land

ASK-219 (pgvector + `study_guide_file_chunks` schema) was deprioritized in favor of the AI pipeline and frontend work. It would have been the foundation for retrieval and citation features that remained in the backlog and were outside the scope of the final deliverable.

## What's Next

This was the final sprint of the project. If development were to continue, the immediate priorities would be:

- Completing the pgvector schema for study guide file chunks
- Wiring the retrieval pipeline to the frontend
- Tackling the AI quiz generation and Q&A panel features that lived in the backlog throughout the project lifecycle

For the full sprint report with issue details, see the [Sprint 3 Review](/docs/sprint-reviews/sprint-3).

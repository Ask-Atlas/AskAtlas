---
slug: sprint-2-recap
title: "Sprint 2: Building the Core Experience"
authors: [askatlas]
tags: [sprint]
---

Sprint 2 focused on building the core user-facing features of AskAtlas now that the foundational infrastructure from Sprint 1 was in place.

<!-- truncate -->

## What We Built

With the infrastructure solid, we shifted to the features users will actually interact with:

- **Dashboard redesign** — Metrics cards, recent study guides list, and activity history sidebar give users an at-a-glance view of their progress.
- **Study Guide View** — Dedicated page displaying guide details, linked quizzes, and referenced resources.
- **Resource Library** — Document upload with drag-and-drop, file management, and list/grid view toggle.
- **File renaming** — PATCH endpoint for updating file metadata with input validation, ownership checks, and full test coverage.

## How We Worked

A recurring theme this sprint was intentional use of mock data on the frontend while backend endpoints are still being developed. This allowed UI work to move forward in parallel without blocking on API availability. Luca led the dashboard and library UI, Nathaniel completed the study guide view, and David contributed the file rename API.

All four pieces of work are in review via open pull requests and pending final merge — no actively worked issues were left incomplete.

## What's Next

Sprint 3 will focus on wiring frontend to real backend data and tackling the CRUD backlog:

- Connect frontend pages to live API endpoints as coverage grows
- Begin course and study guide CRUD operations
- Address code review feedback on open PRs before opening new ones
- Run linters and formatters locally to reduce avoidable CI failures

For the full sprint report with issue details, see the [Sprint 2 Review](/docs/sprint-reviews/sprint-2).

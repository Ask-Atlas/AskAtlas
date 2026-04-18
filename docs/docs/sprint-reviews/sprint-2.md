---
sidebar_position: 2
---

# Sprint 2 Report (03/01/26 – 04/04/26)

[**Sprint demo video**](https://youtu.be/VIqZRLF_-vU)

## What's New (User Facing)

- **Dashboard Redesign** — Dashboard redesign with metrics cards, recent study guides list, and activity history sidebar
- **Study Guide View** — Study Guide View page displaying guide details, linked quizzes, and referenced resources
- **Resource Library** — Library page with document upload, drag-and-drop file management, and list/grid view toggle
- **File Renaming** — File rename capability allowing users to update document names

## Work Summary (Developer Facing)

This sprint focused on building out the core user-facing features of AskAtlas now that the foundational infrastructure from Sprint 1 was in place. Luca led frontend efforts, delivering both the redesigned dashboard and the library UI, which includes drag-and-drop uploads, document previews, and course filtering. Nathaniel completed the study guide view page, implementing a dynamic route with mock data standing in until the study guide API is ready. David contributed on the backend with a new PATCH endpoint for renaming files, complete with input validation, ownership checks, and full test coverage. All four pieces of work are in review via open pull requests and are pending final merge. A recurring theme this sprint was intentional use of mock data on the frontend while backend endpoints are still being developed, which allowed UI work to move forward in parallel without blocking on API availability.

## Unfinished Work

No issues that were actively worked on this sprint were left incomplete. The open pull requests represent finished work currently under code review, not abandoned work. Several backlog issues remain unstarted and have been carried into future sprints as the team continues to prioritize core MVP features.

## Completed Issues/User Stories

| Issue | Description |
|-------|-------------|
| [#6](https://github.com/Ask-Atlas/AskAtlas/issues/6) | UI/UX design for the dashboard |
| [#15](https://github.com/Ask-Atlas/AskAtlas/issues/15) | Design and implement the UI/UX for the Study Guide View |
| [#19](https://github.com/Ask-Atlas/AskAtlas/issues/19) | UI/UX design for the library |
| [#61](https://github.com/Ask-Atlas/AskAtlas/issues/61) | API - PATCH /api/files/{file_id} (update file metadata) |

## Incomplete Issues/User Stories

| Issue | Reason |
|-------|--------|
| [#48](https://github.com/Ask-Atlas/AskAtlas/issues/48) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#47](https://github.com/Ask-Atlas/AskAtlas/issues/47) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#45](https://github.com/Ask-Atlas/AskAtlas/issues/45) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#17](https://github.com/Ask-Atlas/AskAtlas/issues/17) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#14](https://github.com/Ask-Atlas/AskAtlas/issues/14) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#13](https://github.com/Ask-Atlas/AskAtlas/issues/13) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#12](https://github.com/Ask-Atlas/AskAtlas/issues/12) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#11](https://github.com/Ask-Atlas/AskAtlas/issues/11) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#10](https://github.com/Ask-Atlas/AskAtlas/issues/10) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |
| [#9](https://github.com/Ask-Atlas/AskAtlas/issues/9) | We prioritized core UI features this sprint and plan to address this in Sprint 3. |

## Code Files for Review

Please review the following code files, which were actively developed during this sprint, for quality:

| File | Description |
|------|-------------|
| `home/page.tsx` | Redesigned dashboard page with metrics, study guide list, and activity sidebar |
| `study-guide-view.tsx` | Study guide view feature component |
| `resources/page.tsx` | Library document management page |
| `resources/upload/page.tsx` | File upload hub with drag-and-drop |
| `files/service.go` | File rename service with validation and ownership checks |

## Retrospective

### What went well
- Parallel development between frontend UI and backend API worked well this sprint — UI work wasn't blocked waiting on endpoints
- Strong individual ownership over features made progress clear and reduced coordination overhead
- David's backend PR came in with thorough test coverage and clean validation logic from the start
- The team had a clearer sense of scope going into this sprint compared to Sprint 1

### What we'd like to improve
- PRs should be opened and reviewed earlier in the sprint rather than clustering near the deadline
- Some frontend components still have mock data and TODOs that need to be tracked so they don't get forgotten
- CI failures on some PRs (formatting, typecheck) should be caught locally before pushing

### Changes for next sprint
- Begin wiring frontend components to real backend endpoints as API coverage grows
- Address code review feedback on open PRs before opening new ones
- Run linters and formatters locally as a habit before pushing to cut down on avoidable CI failures
- Begin tackling CRUD backlog issues now that UI foundations are in place
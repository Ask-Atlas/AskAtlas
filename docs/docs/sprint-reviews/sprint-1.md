---
sidebar_position: 1
---

# Sprint 1 Report (01/26/26 – 03/01/26)

[**Sprint demo video**](https://youtu.be/3PZzi8RMt3c)

## What's New (User Facing)

- **User Authentication System** — Users can now sign up, log in, and manage their accounts securely through Clerk authentication
- **Homepage and Dashboard** — New landing page showcasing AskAtlas features and an authenticated dashboard for personalized user experience
- **Practice Page UI** — Interface designed for users to engage with practice questions and study materials
- **Responsive Navigation** — Added sidebar navigation for easy access to all authenticated pages

## Work Summary (Developer Facing)

This sprint focused on establishing the foundational infrastructure for AskAtlas. We started from zero and built out the complete authentication pipeline, database architecture, and core UI components.

The team successfully integrated Clerk for authentication with full webhook support to sync user events (creation, updates, deletion) with our PostgreSQL database. We established CI/CD pipelines for both the frontend (Next.js) and backend (Go API), enabling automated deployments.

On the frontend, we implemented the homepage, authenticated dashboard, practice page designs, and a persistent sidebar for navigation. The backend now includes database schemas for users and files, migration scripts, and initial API endpoints for file operations.

A significant challenge was the learning curve. David brought extensive experience with the tech stack and led the infrastructure setup, while Nathaniel and Luca spent considerable time familiarizing themselves with Next.js, Go, PostgreSQL, and the overall system architecture. Despite this, the team successfully delivered a functional foundation with working authentication, deployment pipelines, and initial user-facing features.

## Completed Issues

| Issue | Description |
|-------|-------------|
| [#46](https://github.com/Ask-Atlas/AskAtlas/issues/46) | API - GET /api/files |
| [#44](https://github.com/Ask-Atlas/AskAtlas/issues/44) | Migration - Files Tables |
| [#36](https://github.com/Ask-Atlas/AskAtlas/issues/36) | Add a Sidebar for authenticated pages |
| [#31](https://github.com/Ask-Atlas/AskAtlas/issues/31) | Handle User Deleted Event From Clerk |
| [#30](https://github.com/Ask-Atlas/AskAtlas/issues/30) | Handle User Updated Event From Clerk |
| [#29](https://github.com/Ask-Atlas/AskAtlas/issues/29) | Handle User Created Event From Clerk |
| [#28](https://github.com/Ask-Atlas/AskAtlas/issues/28) | Clerk Configuration For Webhook |
| [#27](https://github.com/Ask-Atlas/AskAtlas/issues/27) | Webhook Endpoint For Clerk |
| [#26](https://github.com/Ask-Atlas/AskAtlas/issues/26) | Database Schema and Setup For Users |
| [#23](https://github.com/Ask-Atlas/AskAtlas/issues/23) | Set up CD for AskAtlas Web Application |
| [#22](https://github.com/Ask-Atlas/AskAtlas/issues/22) | Setup CD for AskAtlas API |
| [#18](https://github.com/Ask-Atlas/AskAtlas/issues/18) | Set up Garage |
| [#16](https://github.com/Ask-Atlas/AskAtlas/issues/16) | Design UI/UX for practice page |
| [#8](https://github.com/Ask-Atlas/AskAtlas/issues/8) | Setting up database tables for Users |
| [#5](https://github.com/Ask-Atlas/AskAtlas/issues/5) | Authenticated Dashboard UI/UX |
| [#4](https://github.com/Ask-Atlas/AskAtlas/issues/4) | Design and Implement UI/UX for the homepage |
| [#3](https://github.com/Ask-Atlas/AskAtlas/issues/3) | Setup Authentication with Clerk on the Frontend |
| [#2](https://github.com/Ask-Atlas/AskAtlas/issues/2) | Boiler plate for frontend |
| [#1](https://github.com/Ask-Atlas/AskAtlas/issues/1) | Boiler plate for AskAtlas API |

All 19 planned issues were completed. Story points were distributed across all team members.

## Planned for Future Sprints

| Issue | Description |
|-------|-------------|
| [#48](https://github.com/Ask-Atlas/AskAtlas/issues/48) | API - DELETE /api/files/:id |
| [#47](https://github.com/Ask-Atlas/AskAtlas/issues/47) | API - PATCH /api/files/:id/visibility |
| [#45](https://github.com/Ask-Atlas/AskAtlas/issues/45) | Migration - Course & Study Guide Join Tables |
| [#19](https://github.com/Ask-Atlas/AskAtlas/issues/19) | UI/UX design for the library |
| [#17](https://github.com/Ask-Atlas/AskAtlas/issues/17) | CRUD for library resources |
| [#15](https://github.com/Ask-Atlas/AskAtlas/issues/15) | Design and implement the UI/UX for the Study Guide View |
| [#14](https://github.com/Ask-Atlas/AskAtlas/issues/14) | CRUD for Quizzes |
| [#13](https://github.com/Ask-Atlas/AskAtlas/issues/13) | Scrape Wazzu for Courses |
| [#12](https://github.com/Ask-Atlas/AskAtlas/issues/12) | CRUD for study guides |
| [#11](https://github.com/Ask-Atlas/AskAtlas/issues/11) | Scrape Wazzu for courses |
| [#10](https://github.com/Ask-Atlas/AskAtlas/issues/10) | CRUD for Courses |
| [#9](https://github.com/Ask-Atlas/AskAtlas/issues/9) | CRUD Operations API Users |
| [#7](https://github.com/Ask-Atlas/AskAtlas/issues/7) | Implement the UI/UX design for the dashboard with Shadcn |
| [#6](https://github.com/Ask-Atlas/AskAtlas/issues/6) | UI/UX design for the dashboard |

## Retrospective

### What went well
- Successfully went from zero to a functional foundation with working auth and deployment pipelines
- Strong initial infrastructure setup with CI/CD for both frontend and backend
- Effective integration of Clerk authentication with complete webhook implementation
- Team collaboration and knowledge sharing between experienced and learning members
- Completed all 19 planned issues for the sprint

### What we'd like to improve
- Onboarding process for team members unfamiliar with the tech stack
- Knowledge distribution — reduce dependency on single team member for infrastructure decisions
- Documentation of setup processes and architectural decisions
- Earlier planning and task breakdown to better estimate complexity

### Changes for next sprint
- Implement pair programming sessions to accelerate learning
- Create internal documentation wiki for common setup procedures and architectural patterns
- Hold brief daily standups to surface blockers earlier
- More granular task breakdown to improve work distribution
- Focus on building out core CRUD operations now that infrastructure is solid

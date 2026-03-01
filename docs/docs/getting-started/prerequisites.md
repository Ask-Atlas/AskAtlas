---
sidebar_position: 1
---

# Prerequisites

This guide outlines the tools needed for your development environment. Windows users should set up WSL first.

## WSL (Windows Only — Do This First)

This project uses makefiles, shell scripts, and toolchains that require a Linux environment. WSL ensures your development environment matches our deployment environment.

📖 [Install WSL](https://learn.microsoft.com/en-us/windows/wsl/install)

## Node.js (via NVM)

Next.js relies on the Node.js runtime. We use NVM to manage Node.js versions and avoid compatibility issues.

📖 [Install NVM on WSL](https://learn.microsoft.com/en-us/windows/dev-environment/javascript/nodejs-on-wsl)

## PNPM

PNPM is our package manager for Node.js dependencies. It's faster, more efficient, and resolves dependencies better than NPM. Used for the web application.

> **Note:** Install PNPM on WSL, not Windows.

📖 [Install PNPM](https://pnpm.io/installation)

## Go

Go is used for our backend API services. Install version **1.24 or higher**.

📖 [Install Go on WSL/Ubuntu](https://dev.to/pu-lazydev/installing-go-golang-on-wsl-ubuntu-18b7)

## Docker

Docker is used to containerize applications for consistent builds across environments. We use it to package images for production deployment.

📖 [Docker Desktop with WSL](https://docs.docker.com/desktop/features/wsl/)

## Infisical CLI

Infisical manages our environment variables. The CLI lets you securely access secrets during development and production, with real-time distribution across the team.

📖 [Install Infisical CLI](https://infisical.com/docs/cli/overview)

## Postman

Postman is used for documenting, testing, and building API endpoints. We use it as a centralized reference for API endpoints across environments.

> **Note:** For Windows, download the native Windows application or use the web version.

- 📖 [Download Postman](https://www.postman.com/downloads/)
- 🔗 [Join our workspace](https://web.postman.co/workspace/6ede04c4-0b85-4121-bbfd-7dc8503262e1)

## DataGrip by JetBrains

DataGrip is a cross-platform database IDE with support for virtually every database. We use it to directly inspect and manage data in our stores.

📖 [Download DataGrip](https://www.jetbrains.com/datagrip/)

## Gemini Student Pro Plan

Google offers nearly unlimited Gemini tokens on the student plan. We can leverage this for developing our AI assistant at zero cost.

📖 [Sign up for Gemini Students](https://gemini.google/students/)

## Optional: nektos/act

Act allows you to run GitHub Actions locally. Useful for writing and testing CI/CD pipelines without pushing to GitHub.

📖 [Install act](https://nektosact.com/installation/)

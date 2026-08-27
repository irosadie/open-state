# Contributing to OpenState

Thank you for your interest in contributing to OpenState! This document outlines
the workflow and expectations for contributors.

## Getting Started

1. **Read the PRD** — [`MAIN_PRD.md`](./MAIN_PRD.md) is the source of truth for
   the product & architecture.
2. **Check the roadmap** — open GitHub issues track planned work by area.
3. **Fork the repository** and create a feature branch.

## Development Setup

Follow the [Quick Start](./README.md#quick-start) to run the stack locally
(Go backend + Next.js frontend + PostgreSQL + Redis).

## How to Contribute

### Reporting Bugs
Open an issue with:
- A clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Environment details (Go/Node versions, OS)

### Suggesting Features
Open an issue with:
- The problem you're solving
- A proposed approach (if any)
- How it relates to the PRD

### Submitting Code

1. Create a branch: `<type>/<short-description>` (e.g. `feature/engine-guard-eval`)
2. Follow the coding conventions:
   - **Frontend**: [Biome](https://biomejs.dev) (no `any`, no `console.*`,
     `const`, double quotes, no semicolons)
   - **Backend**: `go vet`, idiomatic Go, clean architecture
     (domain / application / infrastructure / interfaces)
3. Add tests for new functionality.
4. Run the quality checks:
   ```bash
   # frontend
   cd apps/web && bun run lint && bun run typecheck && bun run test

   # backend
   cd apps/api && go vet ./... && go test ./...
   ```
5. Commit with a clear message, then open a pull request.

## Pull Request Guidelines

- Reference the issue(s) your PR addresses.
- Keep changes focused and minimal.
- Update documentation (README, PRD, ADR) if behavior changes.
- Ensure CI passes (lint, test, build).

## Code of Conduct

Please note that this project adheres to a [Code of Conduct](./CODE_OF_CONDUCT.md).
By participating, you agree to abide by its terms.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](./LICENSE).

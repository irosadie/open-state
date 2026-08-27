---
name: api-bugfix
description: Fix backend bugs with minimal touch, keep other layers stable, then sync DTO, OpenAPI, shared types, tests, and related docs when behavior is affected.
---

# Skill: API Bugfix

## Context (Required)
- Folder scope + impact map: `references/context.md`
- Execution checklist: `templates/checklist.md`

Use this skill when the user asks to fix a backend bug. Core principle: **minimal touch**. Find the root cause, change as little as possible, keep layering clean, then sync DTO, OpenAPI, shared types, and tests when endpoint behavior is actually affected.

## Workflow

### 1. Reproduce or Define the Bug Clearly

Before changing code:
- understand the symptom
- identify the failing behavior in route, controller, service, use case, or repository
- locate the boundary: request binding, business rule, response mapping, repository, or side effect

When possible, add or modify a test that represents the bug.

### 2. Localize the Root Cause

Find the smallest possible root cause. Prioritize by location:
- route / middleware if the bug is in request parsing or auth
- controller if the bug is in request binding or response shaping
- service if the bug is in orchestration
- use case if the bug is business logic
- repository / infra if the bug is in persistence or external side effect

Do not rewrite multiple layers when one layer is enough to fix the bug.

### 3. Apply a Minimal Fix

Minimal touch rules:
- touch as few files as possible
- preserve the `route -> controller -> service -> use case -> repository` boundary
- do not refactor unrelated layers
- do not rename or restructure just because the file is open
- **NEVER** edit sqlc-generated files in `internal/infrastructure/db/` — fix the `.sql` query and regenerate

### 4. Sync Contract and Docs That Are Affected

If the fix changes endpoint behavior, request shape, response shape, or error semantics, update what is actually needed:
- DTO in `internal/application/dtos/`
- `packages/types/` response types
- split OpenAPI in `docs/openapi/`
- relevant tests

Do not let endpoint behavior change while OpenAPI and shared types stay stale.

### 5. Verify Narrowly but Thoroughly

Minimum verification:
- test that reproduces or guards the bug
- `go vet ./...` on the touched surface
- generate OpenAPI if the contract changed
- confirm no new drift between code, shared contract, and docs

## Prohibitions

- **NEVER** refactor across layers when the user's goal is only a bugfix.
- **NEVER** change unrelated files "while you're at it".
- **NEVER** let DTO/OpenAPI/shared types drift when the fix changes endpoint behavior.
- **NEVER** move business logic into the HTTP layer for a quick fix.
- **NEVER** edit sqlc-generated files directly.
- **NEVER** finish without targeted verification.

## Pre-Completion Checklist

- [ ] Bug reproduced or faulty behavior defined clearly
- [ ] Root cause localized to the smallest reasonable layer
- [ ] Changes remain minimal touch
- [ ] DTO/type/docs/OpenAPI updated when affected
- [ ] sqlc regenerated if query changed
- [ ] Relevant tests added or updated
- [ ] `go build ./...` passes
- [ ] No new drift in backend contract
- [ ] All files end with a newline (EOF)

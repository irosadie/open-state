---
name: web-bugfix
description: Fix frontend bugs with minimum-touch changes, keep other surfaces stable, and sync schema, types, hooks, constants, tests, and related docs when impacted.
---

# Skill: Web Bugfix

## Context (Required)
- Folder scope + impact map: `references/context.md`
- Execution checklist: `templates/checklist.md`

Use this skill when the user asks to fix a frontend bug. Core principle: **minimum touch**. Find the root cause, change as little as possible, do not refactor unrelated areas, then sync artifacts that are actually impacted. Leave no drift behind.

## Workflow

### 1. Reproduce or Define the Bug Clearly

Before changing code:
- understand the symptoms
- find reproduction steps or a clear failing behavior
- define the bug's boundary: page, hook, component, shared schema/type, or API integration

If possible, add or update a test that represents the bug.

### 2. Localize the Root Cause

Find the smallest viable root cause. Prioritize these locations:
- `*-page-content.tsx` and its `_components/`
- transaction hooks
- shared schema/type consumed by the frontend
- constants (`api-routers`, `query-keys`)
- reusable components that are actually the source of the bug

Do not refactor large areas just because the bug surfaced in messy code.

### 3. Apply Minimum-Touch Fix

Minimum-touch rules:
- touch as few files as possible
- fix the wrong behavior without redesigning the surrounding area
- do not replace architectural patterns unrelated to the bug
- do not clean up unrelated old debt

### 4. Sync Impacted Artifacts

If the bug touches a contract or data shape, update what truly needs updating:
- `packages/schemas`
- `packages/types`
- frontend transaction hooks
- related constants or utils
- relevant tests
- contract docs if the frontend's expectation of the API genuinely changed

If the bug actually originates from a wrong backend contract, do not hide the problem in the frontend. Patch the frontend surface as needed, then route or continue with the appropriate backend skill.

### 5. Verify Narrowly but Thoroughly

Minimum verification:
- test that reproduces or guards the bug
- typecheck/lint on touched surfaces
- confirm no new contract drift

If the bug lives in an active starter surface, prioritize keeping the homepage, shared contract, and transaction hooks consistent.

## Prohibitions

- **NEVER** do large refactors when the user only asked for a bug fix.
- **NEVER** change other files just because they look messy.
- **NEVER** leave schema/type/hooks/docs drift when the fix changes data shape.
- **NEVER** cover up a backend bug with a frontend workaround that adds state lies or contract mismatch.
- **NEVER** finish without targeted verification.

## Pre-Completion Checklist

- [ ] Bug reproduced or wrong behavior defined clearly
- [ ] Root cause localized to the smallest reasonable surface
- [ ] Change remains minimum touch
- [ ] Shared schema/types/constants/hooks updated if actually impacted
- [ ] Relevant tests added or updated
- [ ] No new drift in the frontend contract
- [ ] All files end with a newline (EOF)

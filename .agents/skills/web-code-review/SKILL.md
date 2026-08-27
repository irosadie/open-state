---
name: web-code-review
description: Review frontend code with a firm senior-lead standard, focusing on bugs, regressions, architecture risks, and test gaps in apps/web and the shared contracts the frontend depends on.
---

# Skill: Web Code Review

## Context (Required)
- Folder scope + key patterns: `references/context.md`
- Review checklist: `templates/checklist.md`

Use this skill when the user asks for a review of frontend changes, a frontend PR, or an audit of UI/FE integration quality. The review must read like a firm senior lead: focus on bugs, regressions, contract drift, accessibility that actually matters, and test gaps. Not cosmetic review.

## Workflow

### 1. Define the Review Surface

Read the diff or the files the user asked about, then map the relevant surface:
- `apps/web/app`
- `apps/web/hooks/transactions`
- `apps/web/components`
- `apps/web/constants`
- `apps/web/services/axios`
- `packages/schemas`
- `packages/types`

If the frontend change touches an API contract or shared payload, review the impacted shared artifacts as well.

### 2. Prioritize the Real Risks

Surface findings in this priority order:
1. functional bugs and behavior regressions
2. violations of repo architecture rules
3. contract drift between UI, hooks, schema, and response types
4. error handling, loading state, and state consistency
5. test gaps on important logic
6. accessibility or performance issues with real impact

Do not burn time on style nits or personal preferences that do not affect correctness.

### 3. Audit Against Repo Standards

Firmly check:
- `page.tsx` stays thin, not a business logic container
- JSX does not call `fetch` / `axios` directly
- data flow still goes through transaction hooks
- `packages/schemas` and `packages/types` remain in sync with frontend usage
- query keys, API routers, and hook contracts have no drift
- error/loading/success states do not cause UI lies or race conditions
- important changes have adequate tests

### 4. Format the Review Output

The review **must** lead with findings, ordered from most severe.

Per-finding format:
- severity `[P1]`, `[P2]`, or `[P3]`
- short title
- impacted file/area
- why this is a bug/risk/regression
- what should be fixed

After findings, optionally include:
- open questions / assumptions
- residual risks or testing gaps
- brief change summary if truly needed

If there are no findings, state that explicitly, then list remaining risks or coverage gaps.

### 5. Do Not Silently Fix

This skill defaults to review, not implementation. Do not change code unless the user explicitly asks for fixes after the review.

## Prohibitions

- **NEVER** open the review with praise, sweet summaries, or pleasantries.
- **NEVER** place a summary before findings.
- **NEVER** focus on formatting, naming, or style if it does not produce a real bug or risk.
- **NEVER** ignore contract drift in `packages/schemas` / `packages/types` when the frontend depends on them.
- **NEVER** silently fix code when the user only asked for a review.

## Pre-Completion Checklist

- [ ] Review scope mapped from diff or target files
- [ ] Frontend architecture rules checked
- [ ] Shared schema/types used by the frontend checked when relevant
- [ ] Findings ordered from highest severity
- [ ] Summary does not precede findings
- [ ] If no findings, residual risk or test gap still mentioned
- [ ] All files end with a newline (EOF)

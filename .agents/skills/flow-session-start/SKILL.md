---
name: flow-session-start
description: Handle Start or Mulai commands for repo onboarding: check MCP, registry, memory, and work status then direct user to the appropriate next step.
---

# Skill: Session Start

## Context (Required)
- Folder scope + quick-check paths: `references/context.md`
- Execution checklist: `templates/checklist.md`

Use this skill when the user only types commands like `Mulai`, `Start`, `Mulai Vibe Coding`, or similar variants without a specific technical task. The purpose of this skill is to transform that short command into a fast, interactive onboarding session relevant to the current repo state.

## Workflow

### 1. Recognize Start Trigger

Activate this skill if the user:
- only writes `Mulai`, `Start`, `Mulai Vibe Coding`, or equivalent variants
- asks to start a work session from scratch
- asks the agent to check repo readiness before continuing

If the user has already given a specific implementation task, **do not** use this skill as the main workflow.

### 2. Run Quick Status Check

Before asking anything, read `.agents/settings.json` then run the status script:

```bash
bun run session:status
```

Summarize the results into four areas:
- MCP status (`.mcp.json`, required servers, missing servers)
- OpenSpec changes status (`openspec/changes/`)
- Serena memory status (check if onboarding completed)
- workspace status (`git branch`, worktree dirty/clean)

### 3. Classify Session State

Use the following heuristics:
- If MCP is not complete, treat this as **likely first-time init**
- If branch is not `main` or worktree is dirty, treat this as **likely resume last task**
- If workspace is clean and MCP is ready, treat this as **ready to start new work item**

### 4. Present a Short Summary

Reply with a brief summary, for example:
- whether MCP is ready or not
- what features are in the registry
- whether there are signs of work in progress

Do not dump raw JSON to the user unless requested.

### 5. Direct to the Appropriate Next Step

After the summary, direct to one of the following paths:
- MCP not ready → continue to `ops-mcp-setup`
- work likely not finished → offer to resume last task with `openspec-apply-change`
- repo ready and user wants to start something new → continue to `openspec-propose`

Ask **one clear next step question**, not a long list of questions.

Example patterns:
- `MCP for GitHub, Jira, and Notion is not complete. Should I set up MCP first?`
- `Workspace shows an unfinished work branch. Should I continue the last task from this branch?`
- `Repo is clean and MCP is ready. Should we break down a new feature now?`

### 6. Don't Jump to Implementation

For start commands that are still general:
- don't immediately write code
- don't immediately create proposals or specs
- don't immediately set up MCP tokens without user confirmation

This skill stops after the session status is clear and the user chooses the next step.

## Prohibitions

- **FORBIDDEN** to respond to `Mulai` with a generic answer that does not check the repo.
- **FORBIDDEN** to immediately ask many discovery questions before running the quick-check.
- **FORBIDDEN** to assume MCP is ready without checking; verify `.mcp.json` and required servers first.
- **FORBIDDEN** to immediately write `.mcp.json` with placeholder tokens.
- **FORBIDDEN** to immediately start feature implementation without confirming the next path with the user.

## Pre-Completion Checklist

- [ ] Trigger is appropriate for onboarding start session
- [ ] `.agents/settings.json` read
- [ ] `bun run session:status` executed
- [ ] MCP, registry, memory, and workspace briefly summarized
- [ ] Session state classified: first init, resume, or new work
- [ ] User given one clear next step question
- [ ] If needed, next step directed to correct skill (`ops-mcp-setup`, `openspec-propose`, or `openspec-apply-change`)
- [ ] All files end with a newline (EOF)

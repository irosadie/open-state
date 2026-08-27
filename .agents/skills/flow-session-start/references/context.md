# Context: Session Start

## Target Folder

```
.agents/settings.json                → source of truth trigger + onboarding config
openspec/changes/                   → active OpenSpec changes
scripts/start-session.mjs           → quick status check for start session
.mcp.json                           → indicates whether MCP repo-level is configured or not
```

Note: Serena memories are the system for persistent context across sessions. Check if Serena onboarding completed.

## Required Script

Run this when the skill is used:

```bash
bun run session:status
```

Script output will summarize:
- MCP status required vs configured
- active OpenSpec changes
- Serena memory status
- active branch + number of changed files

## Important Patterns

- For first init: direct to skill `ops-mcp-setup`
- For new feature: direct to OpenSpec `/opsx:propose`
- For continuing existing task: direct to OpenSpec `/opsx:apply`

## Good Summary

Summary format to user should be short and operational. Simply convey:
1. Whether MCP is ready or not
2. Active OpenSpec changes
3. Whether there are signs of work in progress or not
4. One next step question

Avoid:
- dumping raw JSON
- asking many questions at once
- immediately starting implementation

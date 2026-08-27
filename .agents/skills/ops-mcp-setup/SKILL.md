---
name: ops-mcp-setup
description: Configure the GitHub MCP server in this project so the agent can interact with the repository. Use on first-time project setup or when MCP is not yet connected.
---

# Skill: Ops MCP Setup

## Context (Required)
- Token details + file format: `references/context.md`
- Execution checklist: `templates/checklist.md`

Connect the agent to the **GitHub** MCP so it can read issues, create PRs, and interact with the repository.

## Prerequisites

- `npx` available in PATH
- Access to a GitHub account with the target repo

## Workflow

### 1. Ask for Required Information

Before creating files, ask the user:

```
1. GitHub Personal Access Token (scope: repo, read:org)
```

Do not create `.mcp.json` with placeholders — value must be a real token from the user.

### 2. Check `.gitignore`

Ensure `.mcp.json` is in `.gitignore`. If not, add:

```
.mcp.json
```

### 3. Create `.mcp.json`

Create the file at project root with the confirmed token:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "{token from user}"
      }
    }
  }
}
```

### 4. Update `.agents/settings.json`

After MCP is ready, update the `repo` section in `.agents/settings.json`:

```json
{
  "repo": {
    "owner": "{github org/username}",
    "name": "{github repo name}"
  }
}
```

### 5. Verify

Instruct the user to:

1. **Restart Claude Code / agent** so `.mcp.json` is loaded
2. Test connection: "List open issues in repo {repo-name}"

If it fails, surface the error and direct to the troubleshooting step in `references/context.md`.

## Prohibitions

- **FORBIDDEN** to store token in any file other than `.mcp.json`.
- **FORBIDDEN** to commit `.mcp.json` — ensure it is in `.gitignore`.
- **FORBIDDEN** to create `.mcp.json` with placeholder values.

## Pre-Completion Checklist

- [ ] `.mcp.json` is in `.gitignore`
- [ ] Token confirmed from user (not a placeholder)
- [ ] `.mcp.json` created at project root with `github` MCP server
- [ ] `.agents/settings.json` updated: `repo.owner`, `repo.name`
- [ ] User instructed to restart agent
- [ ] User successfully verified GitHub connection

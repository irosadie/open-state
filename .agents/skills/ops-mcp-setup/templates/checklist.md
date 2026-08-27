# Checklist: Ops MCP Setup

## Preparation

- [ ] Read `references/context.md` (how to obtain token)
- [ ] Confirm user has access to GitHub repo

## Collect Token from User

- [ ] **GitHub PAT** confirmed (starts with `ghp_`, scopes: `repo`, `read:org`, `read:user`)
- [ ] **GitHub repo** confirmed (owner + repo name)

## Create / Update Files

- [ ] Check `.gitignore` — ensure `.mcp.json` is listed. If not, add it
- [ ] `.mcp.json` created at project root with `github` server
- [ ] Values in `.mcp.json` are real tokens (no placeholders)
- [ ] `.agents/settings.json` updated: `repo.owner`, `repo.name`

## Verification

- [ ] User instructed to restart agent
- [ ] GitHub MCP verified: able to list issues or repos
- [ ] If errors occur, consult troubleshooting table in `references/context.md`

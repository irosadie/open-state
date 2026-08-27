# Context: Ops MCP Setup

## Files Created / Modified

```
.mcp.json              → MCP config (DO NOT commit)
.gitignore             → ensure .mcp.json is listed here
.agents/settings.json  → update repo.owner + repo.name
```

## `.mcp.json` Format

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxxxxxxxxxxxxxxxxxxx"
      }
    }
  }
}
```

---

## How to Get Token: GitHub

**URL:** https://github.com/settings/tokens → Personal access tokens → Tokens (classic)

**Steps:**
1. Click **Generate new token (classic)**
2. Name it (e.g. `claude-code-mcp`)
3. Set expiration as needed
4. Check scopes: `repo`, `read:org`, `read:user`
5. Click **Generate token**
6. Copy token (shown only once, starts with `ghp_`)

**Example token:** `ghp_AbCdEfGhIjKlMnOpQrStUvWxYz123456`

---

## Troubleshooting

| Error | Likely Cause | Fix |
|---|---|---|
| GitHub: `401 Unauthorized` | Token wrong or expired | Generate new token in GitHub settings |
| MCP not appearing in agent | `.mcp.json` not loaded | Fully restart agent |

---

## Sections in `settings.json` to Update

```json
{
  "repo": {
    "owner": "my-github-org",
    "name": "my-repo"
  }
}
```

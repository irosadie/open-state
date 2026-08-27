# Context: Skill Creator

## Target Folder

```
.agents/skills/{skill-name}/
├── SKILL.md
├── references/
│   └── context.md
├── templates/
│   └── checklist.md
└── agents/
    └── openai.yaml         → Codex/OpenAI metadata

.claude/skills/{skill-name}/
└── SKILL.md              → Claude wrapper (pointer to source of truth)
```

## Files Modified

```
.agents/skills/manifest.json   → add new skill entry
.agents/AGENTS.md              → add to Skill Registry + SKILL references
CLAUDE.md                      → add to Skill Registry table
```

## Commands

```bash
bun run skills:create -- --name {scope}-{capability} --scope {scope} --description "{description}" --when "{when used}"
bun run skills:sync-registry
bun run skills:sync-claude
bun run skills:sync
bun run skills:validate
```

## Naming Convention

| Component | Format | Example |
|---|---|---|
| Skill folder | `{scope}-{capability}` | `web-slicing`, `ops-docker` |
| Scope | `frontend`, `backend`, `docs`, `ops`, `flow`, `meta` | `meta` |
| Description | Active sentence, when used + output | "Create a new skill from scratch..." |

## Available Scopes

| Scope | Used For |
|---|---|
| `frontend` | Skills that touch `apps/web/` |
| `backend` | Skills that touch `apps/api/` or `apps/worker/` |
| `docs` | Skills that produce documentation |
| `ops` | Infra/deploy skills (Dockerfile, CI/CD) |
| `flow` | Orchestration skills (GitHub, OpenSpec) |
| `meta` | Skills that govern the agent system itself |

## `SKILL.md` Pattern (Minimal Template)

```markdown
---
name: {skill-name}
description: {One-sentence description}
---

# Skill: {Title}

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

{Purpose paragraph}

## Workflow

1. ...
2. ...

## Prohibitions

- **NEVER** ...

## Pre-Completion Checklist

- [ ] ...
- [ ] All files end with a newline (EOF)
```

## Claude Wrapper Pattern (`.claude/skills/{name}/SKILL.md`)

```markdown
---
name: "{skill-name}"
description: "{One-sentence description}"
---

Source of truth lives at `.agents/skills/{skill-name}/SKILL.md`.

When this skill is used:
1. Read `.agents/skills/{skill-name}/SKILL.md`.
2. Follow the workflow and rules in that file.
3. Read referenced child files (`references/context.md`, `templates/checklist.md`) from the source-of-truth folder.
```

## `agents/openai.yaml` Pattern

```yaml
interface:
  display_name: "{Short Skill Title}"
  short_description: "{Short description for the skill picker}"
  default_prompt: "Use $skill-name to {main outcome of this skill}."
policy:
  allow_implicit_invocation: true
```

## `manifest.json` Entry Pattern

```json
{
  "name": "{skill-name}",
  "description": "{One-sentence description}",
  "scope": "{scope}",
  "whenToUse": "{Skill Registry text}",
  "path": ".agents/skills/{skill-name}"
}
```

## Existing Skills (Reference)

| Skill | Scope | See |
|---|---|---|
| `web-slicing` | frontend | `.agents/skills/web-slicing/` |
| `api-feature` | backend | `.agents/skills/api-feature/` |
| `docs-openapi` | docs | `.agents/skills/docs-openapi/` |

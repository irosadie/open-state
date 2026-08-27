# OpenCode Instructions

@.agents/AGENTS.md

Source of truth for skills is `.agents/skills/`.
Read `.agents/settings.json` at the start of each task.

## Skills (OpenCode)

Skills are registered in `.opencode/skill/`. Each skill wraps its source of
truth in `.agents/skills/`.

To use a skill, read its `SKILL.md` in `.opencode/skill/<name>/`, then follow
the source of truth in `.agents/skills/<name>/` (including `references/context.md`
and `templates/checklist.md`).

Available skills: openspec-propose, openspec-apply-change, openspec-archive-change,
openspec-explore, openspec-sync-specs, api-feature, api-bugfix, api-code-review,
db-sqlc-schema, docs-openapi, web-slicing, web-api-integrated, web-bugfix,
web-code-review, web-seo-geo-friendly, ops-docker, ops-mcp-setup,
skill-creator, skill-add-example, meta-skill-hygiene.

## OpenSpec Commands

Use OpenSpec via slash commands. Each maps to the corresponding skill:

- `/opsx:propose` — turn an idea into proposal + specs + design + tasks
  (skill: `openspec-propose`)
- `/opsx:apply` — implement tasks from an active OpenSpec change
  (skill: `openspec-apply-change`)
- `/opsx:archive` — archive a completed change (skill: `openspec-archive-change`)
- `/opsx:explore` — explore/think through an idea before committing
  (skill: `openspec-explore`)
- `/opsx:sync` — sync delta specs to main specs (skill: `openspec-sync-specs`)

OpenSpec artifacts live in `openspec/changes/<change-id>/`
(proposal.md, specs/, design.md, tasks.md).

## Start Session

If user types `Start`, `Mulai`, or `Mulai Vibe Coding`:
1. Run `bun run session:status`
2. Summarize MCP status, branch, and worktree
3. Classify: first init, resume last task, or ready for new work
4. Ask one clear next-step question

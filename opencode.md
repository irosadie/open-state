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

Available skills:
<!-- skill-links:start -->
- `web-api-integrated` — Integrate API endpoint to frontend with schema, types, constants, and hooks
- `web-bugfix` — Fix frontend bug with minimal touch and sync impacted contracts
- `web-code-review` — Review frontend code strictly before merge or during quality audit
- `web-seo-geo-friendly` — SEO and GEO optimization for public Next.js pages
- `web-slicing` — Implement UI from design, screenshot, or Figma
- `api-bugfix` — Fix backend bug with minimal touch and sync impacted contracts
- `api-code-review` — Review backend code strictly before merge or during quality audit
- `api-feature` — Implement new backend feature following Clean Architecture
- `db-sqlc-schema` — Changes to goose migrations, sqlc queries, and PostgreSQL schema validation
- `docs-openapi` — Write or update split OpenAPI documentation per feature
- `ops-docker` — Write or modify backend Dockerfile for Linux deployment
- `ops-mcp-setup` — Setup GitHub MCP for this repo's workflow
- `meta-skill-hygiene` — Audit and maintain skill metadata consistency
- `skill-add-example` — Add reusable example code for other skills
- `skill-creator` — Create or update skills with consistent format
- `openspec-apply-change` — Implement tasks from an active OpenSpec change
- `openspec-archive-change` — Archive a completed OpenSpec change
- `openspec-explore` — Explore and think through ideas before committing to a change
- `openspec-propose` — Start a new feature — turn an idea into proposal + specs + design + tasks
- `openspec-sync-specs` — Sync delta specs to main specs
<!-- skill-links:end -->

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

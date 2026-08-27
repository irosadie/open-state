# Claude Project Instructions

@.agents/AGENTS.md

Claude-compatible skill entries are in `.claude/skills/`.
Source of truth remains in `.agents/skills/`.

## Additional Principles for Claude

- **Always apply current best practices** — use recommended patterns and APIs for every technology (Next.js App Router, Hono, BullMQ, Prisma, React Query, Zod). Don't use old patterns when better ones exist.
- **Search the web if unsure** — if you don't know the latest approach, correct API, or best practice, **use WebSearch or WebFetch before writing code**. Better slow and correct than fast and wrong.
- **Follow the established flow** — planning via OpenSpec (`/opsx:propose`), then implementation per task with the matching skill. Implementation order per feature: Slicing → Backend+OpenAPI → FE↔API Integration.
- **Read skill before executing** — every task has its skill. Read `SKILL.md` + `references/context.md` + `templates/checklist.md` before starting.
- **Follow Biome rules** — before writing code, read `biome.json` at root. Forbidden: `any`, `console.*`, unused variables/imports. Required: `const`, double quote, no semicolons. Your code must pass all rules.

## Start Session

If the user only types `Start`, `Mulai`, `Mulai Vibe Coding`, or similar:
1. Run `bun run session:status`
2. Summarize MCP status, branch, and worktree
3. Determine if this is first init, resume last task, or ready for new work
4. Ask one clear next-step question


---

## Skill Registry

<!-- skill-registry:start -->
| Skill | Scope | When to Use |
|---|---|---|
| `web-api-integrated` | Frontend | Integrate API endpoint to frontend with schema, types, constants, and hooks |
| `web-bugfix` | Frontend | Fix frontend bug with minimal touch and sync impacted contracts |
| `web-code-review` | Frontend | Review frontend code strictly before merge or during quality audit |
| `web-seo-geo-friendly` | Frontend | SEO and GEO optimization for public Next.js pages |
| `web-slicing` | Frontend | Implement UI from design, screenshot, or Figma |
| `api-bugfix` | Backend | Fix backend bug with minimal touch and sync impacted contracts |
| `api-code-review` | Backend | Review backend code strictly before merge or during quality audit |
| `api-feature` | Backend | Implement new backend feature following Clean Architecture |
| `db-sqlc-schema` | Backend | Changes to goose migrations, sqlc queries, and PostgreSQL schema validation |
| `docs-openapi` | Docs | Write or update split OpenAPI documentation per feature |
| `ops-docker` | Ops | Write or modify backend Dockerfile for Linux deployment |
| `ops-mcp-setup` | Ops | Setup GitHub MCP for this repo's workflow |
| `meta-skill-hygiene` | Meta | Audit and maintain skill metadata consistency |
| `skill-add-example` | Meta | Add reusable example code for other skills |
| `skill-creator` | Meta | Create or update skills with consistent format |
| `openspec-apply-change` | OpenSpec | Implement tasks from an active OpenSpec change |
| `openspec-archive-change` | OpenSpec | Archive a completed OpenSpec change |
| `openspec-explore` | OpenSpec | Explore and think through ideas before committing to a change |
| `openspec-propose` | OpenSpec | Start a new feature — turn an idea into proposal + specs + design + tasks |
| `openspec-sync-specs` | OpenSpec | Sync delta specs to main specs |
<!-- skill-registry:end -->

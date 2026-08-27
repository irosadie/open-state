# Checklist: Skill Creator

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Check `.agents/skills/manifest.json` — confirm the skill does not already exist
- [ ] Define skill name: kebab-case, format `{scope}-{capability}`
- [ ] Define scope: `frontend` / `backend` / `docs` / `ops` / `flow` / `meta`
- [ ] Write a one-sentence description: what it does + when used

## Create Skill Files (`.agents/skills/{name}/`)

- [ ] `SKILL.md` created with frontmatter (`name`, `description`)
- [ ] `SKILL.md` has sections: Context, Workflow, Prohibitions, Checklist
- [ ] `references/context.md` created — target folder + key patterns + technology table
- [ ] `templates/checklist.md` created — all steps ordered and checkable
- [ ] `agents/openai.yaml` created — Codex/OpenAI metadata (`interface`, `policy`, `dependencies` if needed)

## Create Claude Wrapper (`.claude/skills/{name}/`)

- [ ] `.claude/skills/{name}/SKILL.md` created
- [ ] Contents: frontmatter + pointer to source of truth + 3-step instructions

## Registration

- [ ] Entry added in `.agents/skills/manifest.json` (`name`, `description`, `scope`, `path`)
- [ ] Row added to Skill Registry table in `.agents/AGENTS.md`
- [ ] Reference link added in SKILL section of `.agents/AGENTS.md`
- [ ] Row added to Skill Registry table in `CLAUDE.md`
- [ ] Run `bun run skills:create -- --help` if you need the correct scaffold command
- [ ] `whenToUse` field written in `manifest.json`
- [ ] Run `bun run skills:sync`
- [ ] Run `bun run skills:validate`

## Final Validation

- [ ] Skill name consistent across all files (SKILL.md, Claude wrapper, manifest)
- [ ] Description consistent across `SKILL.md`, Claude wrapper, and manifest
- [ ] `agents/openai.yaml` does not duplicate the main skill instructions
- [ ] No typos in reference paths
- [ ] All files end with a newline (EOF)

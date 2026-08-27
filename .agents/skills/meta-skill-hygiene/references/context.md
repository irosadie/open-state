# Context: Meta Skill Hygiene

## Target Files

- `.agents/skills/*/SKILL.md` → source of truth for skill name, description, workflow
- `.agents/skills/*/agents/openai.yaml` → Codex/OpenAI metadata
- `.agents/skills/manifest.json` → registry source of truth (`scope`, `path`, `whenToUse`)
- `.agents/scripts/` → automation for create/sync/validate
- `.claude/skills/*/SKILL.md` → generated wrapper from the source of truth
- `.agents/AGENTS.md` and `CLAUDE.md` → generated registry sections via markers

## Commands

```bash
bun run skills:create -- --name {scope-capability} --scope {scope} --description "{description}" --when "{whenToUse}"
bun run skills:sync
bun run skills:sync-registry
bun run skills:sync-claude
bun run skills:validate
```

## Key Patterns

- Registry changes must originate from `manifest.json`, never manual table edits.
- The Claude wrapper must always be regenerated from `SKILL.md`.
- `openai.yaml` holds interface/policy metadata only — never the main skill instructions.
- Run the validator before merging changes under `.agents/` or `.claude/`.

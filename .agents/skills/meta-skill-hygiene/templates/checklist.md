# Checklist: Meta Skill Hygiene

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Identify which source-of-truth file changed (`SKILL.md`, `manifest.json`, `openai.yaml`)

## Execution

- [ ] Core edits go into the source of truth, not the generated output
- [ ] Run `bun run skills:sync`
- [ ] If only one output needs regeneration, run the matching specific sync command

## Finalization

- [ ] Run `bun run skills:validate`
- [ ] No drift between `manifest.json`, the Claude wrapper, and the registry
- [ ] No placeholders or broken marker sections
- [ ] Every file ends with a newline (EOF)

---
name: meta-skill-hygiene
description: Audit, sync, and maintain skill metadata consistency across this repo. Use when skills, the registry, Claude wrappers, or agent metadata change.
---

# Skill: Meta Skill Hygiene

## Context (Required)
- Scope + target files: `references/context.md`
- Execution checklist: `templates/checklist.md`

Use this skill to keep every skill asset in sync: `SKILL.md`, `manifest.json`, `agents/openai.yaml`, the Claude wrapper, and the registry sections in `AGENTS.md` / `CLAUDE.md`.

## Workflow

1. Identify which source of truth changed:
   - `SKILL.md` for name and description
   - `manifest.json` for `scope`, `path`, and `whenToUse`
   - `agents/openai.yaml` for Codex/OpenAI metadata
2. For new skills or renames, scaffold/update the source of truth first.
3. Run `bun run skills:sync` to regenerate the registry and Claude wrapper.
4. Run `bun run skills:validate` to confirm no drift or missing files.
5. If validation fails, fix the source of truth, then sync and validate again.

## Prohibitions

- **NEVER** treat the Claude wrapper or the generated registry sections as the source of truth.
- **NEVER** edit blocks between `<!-- skill-registry:* -->` or `<!-- skill-links:* -->` markers without re-running sync.
- **NEVER** fix drift by editing the generated output only; fix the source file.
- **NEVER** leave `manifest.json` without `whenToUse` for any skill.

## Pre-Completion Checklist

- [ ] Source of truth for the change is clear
- [ ] `bun run skills:sync` run
- [ ] `bun run skills:validate` passes
- [ ] No drift between manifest, wrapper, and registry
- [ ] Every file ends with a newline (EOF)

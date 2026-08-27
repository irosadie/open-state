# OpenCode Instructions

@.agents/AGENTS.md

Source of truth for skills is `.agents/skills/`.
Read `.agents/settings.json` at the start of each task.

## Skill Invocation

When a matching skill exists in `.agents/skills/`, read that `SKILL.md` plus its `references/context.md` and `templates/checklist.md` before executing.

## Start Session

If user types `Start`, `Mulai`, or `Mulai Vibe Coding`:
1. Run `bun run session:status`
2. Summarize MCP status, branch, and worktree
3. Classify: first init, resume last task, or ready for new work
4. Ask one clear next-step question

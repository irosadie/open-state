# Checklist: Session Start

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Validate that the user trigger is indeed a start/onboarding command

## Quick Check

- [ ] Run `bun run session:status`
- [ ] Summarize MCP status
- [ ] Summarize active OpenSpec changes
- [ ] Summarize memory status
- [ ] Summarize branch and worktree

## Direct Next Step

- [ ] Determine whether this is first init, resume, or new work
- [ ] If MCP is not ready, direct to `ops-mcp-setup`
- [ ] If there is work in progress, offer to continue last task with `openspec-apply-change`
- [ ] If repo is ready for new task, direct to `openspec-propose`
- [ ] Ask one clear next step question

## Finalization

- [ ] No false assumptions about MCP or last task
- [ ] No feature implementation started without user confirmation
- [ ] All files end with a newline (EOF)

# Checklist: Web Bugfix

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Write down the bug symptoms or reproduction steps

## Execution

- [ ] Root cause localized to the smallest file/area
- [ ] Change remains minimum touch
- [ ] Shared schema/types/hooks/constants updated when actually impacted
- [ ] No unrelated cleanup/refactor
- [ ] Reproduction or guard test added/updated

## Finalization

- [ ] Frontend contract stays in sync
- [ ] Relevant lint/type/test run
- [ ] No unrelated changes snuck in
- [ ] All files end with a newline (EOF)

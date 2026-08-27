# Checklist: Web Code Review

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Identify the frontend files and shared contracts to review

## Core Review

- [ ] Look for behavior bugs/regressions first
- [ ] Check frontend architecture rules
- [ ] Check schema/types/constants/hooks drift when relevant
- [ ] Check loading/error/success states that could mislead the UI
- [ ] Check test gaps on important logic

## Output

- [ ] Findings written first, ordered by severity
- [ ] No irrelevant stylistic nits
- [ ] Open questions or residual risks written after findings
- [ ] If no findings, state that explicitly
- [ ] All files end with a newline (EOF)

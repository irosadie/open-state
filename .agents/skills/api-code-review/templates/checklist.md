# Checklist: API Code Review

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Identify backend files and contract artifacts to review

## Core Review

- [ ] Hunt endpoint bugs/regressions first
- [ ] Check clean architecture layering
- [ ] Check validator/DTO/schema/type/OpenAPI drift when relevant
- [ ] Check error handling and status code behavior
- [ ] Check test gaps for important behavior

## Output

- [ ] Findings written first, ordered by severity
- [ ] No irrelevant stylistic nits
- [ ] Open questions or residual risks written after findings
- [ ] If no findings, state that explicitly
- [ ] All files end with a newline (EOF)

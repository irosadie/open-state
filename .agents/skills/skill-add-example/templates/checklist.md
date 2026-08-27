# Checklist: Skill Add Example

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Identify target skill that will use this example
- [ ] Pick example name (kebab-case, descriptive)
- [ ] Check whether a similar example exists in `.agents/examples/{skill-name}/`

## Choose Location

- [ ] Skill folder: `.agents/examples/{skill-name}/`
- [ ] Context subfolder determined (framework/domain, e.g. `nextjs-app-router`, `hooks`)
- [ ] Example folder created: `.agents/examples/{skill-name}/{subfolder}/{example-name}/`

## Write Example Files

- [ ] Code sourced from real codebase or established pattern
- [ ] Specific business logic replaced with clear placeholders
- [ ] No secrets, tokens, or sensitive data
- [ ] First-line comment: `// Example: {short description}`
- [ ] File structure follows related skill's convention

## Update Skill References

- [ ] Open `references/context.md` in related skill
- [ ] Add entry `### Example N — {Name}` with: location, pattern, file list, when to use
- [ ] Example numbers are sequential (no gaps)
- [ ] Update `SKILL.md` in related skill if it explicitly references examples

## Final Validation

- [ ] Example path consistent across all references
- [ ] Files directly usable as templates by agent
- [ ] All files end with newline (EOF)

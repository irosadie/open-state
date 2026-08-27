---
name: skill-add-example
description: Add real code examples to .agents/examples/ so other skills can reference them. Use when a new pattern needs to be standardized as a template or agent reference.
---

# Skill: Skill Add Example

## Context (Required)
- Scope folder + conventions: `references/context.md`
- Execution checklist: `templates/checklist.md`

Add proven, real code examples to `.agents/examples/` so the agent can reference them when executing skills.

## Workflow

### 1. Determine Scope & Target Skill
- Identify the skill that will use this example (e.g. `web-slicing`, `web-api-integrated`)
- Pick a descriptive subfolder name (kebab-case): e.g. `use-products`

### 2. Choose Example Location

```
.agents/examples/
└── {skill-name}/           → folder per skill
    └── {subfolder}/        → specific context (framework, domain, etc.)
        └── {example-name}/ → example name (kebab-case, descriptive)
            ├── {file}.ts   → real code file
            └── {file}.tsx  → real code file
```

Existing locations:
- `.agents/examples/web-slicing/nextjs-app-router/examples/` — basic CRUD
- `.agents/examples/web-api-integrated/hooks/use-examples/` — CRUD hooks

### 3. Write Example Files

- Copy real code from the codebase (not fabricated) — if none exists, build from an established pattern
- Ensure the code is **directly usable as a template** by the agent
- Replace overly specific business logic with clear placeholders
- Preserve original variable names and structure for clarity

### 4. Update References in Related Skill

After adding the example, update `references/context.md` in the skill that uses it:

```markdown
## Real Code Examples

### Example N — {Example Name}
Location: `.agents/examples/{skill-name}/{subfolder}/{example-name}/`

Pattern: {list of patterns demonstrated}

Files:
- `{file}.tsx` — {short description}

**Use for:** {scenario when this example is relevant}
```

### 5. Make Example Discoverable

- Add a comment on the first line of the file: `// Example: {short description}`
- For multi-file examples, add a brief `README.md` in the example folder (optional, only if complex)

## Prohibitions

- **FORBIDDEN** to add examples that have not been used in real codebases — validate in the project first.
- **FORBIDDEN** to store secrets, credentials, or sensitive data in example files.
- **FORBIDDEN** to create examples at the root `.agents/examples/` without a skill subfolder.
- **FORBIDDEN** to duplicate existing examples — check first.
- **FORBIDDEN** to skip updating `references/context.md` in the related skill.

## Pre-Completion Checklist

- [ ] Scope and target skill identified
- [ ] Folder location chosen per convention (`{skill-name}/{subfolder}/{example-name}/`)
- [ ] Example files written — real code, usable as template
- [ ] No secrets or sensitive data in code
- [ ] `references/context.md` in related skill updated with new example entry
- [ ] All files end with newline (EOF)

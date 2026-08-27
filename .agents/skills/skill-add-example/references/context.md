# Context: Skill Add Example

## Target Folder

```
.agents/examples/
├── web-slicing/
│   └── nextjs-app-router/
│       └── examples/              → basic CRUD template
└── web-api-integrated/
    └── hooks/
        └── use-examples/          → hook folder example per domain
```

## Example Folder Structure

```
.agents/examples/{skill-name}/{subfolder}/{example-name}/
├── {feature}-content.tsx   → Client Component (for web-slicing)
├── page.tsx                → Suspense wrapper (for web-slicing)
├── use-data-table.ts       → hook (for web-api-integrated)
├── use-get-one.ts
├── use-insert-one.ts
├── use-update-one.ts
├── use-delete-one.ts
└── index.ts
```

## Naming Convention

| Component | Format | Example |
|---|---|---|
| Skill folder | Exactly matches skill name | `web-slicing` |
| Context subfolder | kebab-case, framework/domain | `nextjs-app-router`, `hooks` |
| Example name | kebab-case, domain/feature name | `use-products` |
| File | per codebase convention | `use-products-content.tsx` |

## Skill → Examples Mapping

| Skill | Examples Folder |
|---|---|
| `web-slicing` | `.agents/examples/web-slicing/nextjs-app-router/` |
| `web-api-integrated` | `.agents/examples/web-api-integrated/hooks/` |
| New skill | `.agents/examples/{skill-name}/` |

## How to Update `references/context.md` in Related Skill

Find the `## Real Code Examples` section in the related skill's `references/context.md`, then add:

```markdown
### Example N — {Descriptive Name}
Location: `.agents/examples/{skill-name}/{subfolder}/{example-name}/`

Pattern: {pattern-1} + {pattern-2} + {pattern-3}

Files:
- `{file}.tsx` — {file purpose}
- `{file2}.ts` — {file purpose}

**Use for:** {specific condition when this example is relevant}

---
```

## Existing Entry Reference

### Entry in `web-slicing/references/context.md`

```markdown
### Example 1 — CRUD Basic (Template)
Location: `.agents/examples/web-slicing/nextjs-app-router/examples/`

Pattern: useDataTable + useQueryParam + debounce + SweetAlert preConfirm + ActionsDropdown

Files:
- `page.tsx` — Suspense wrapper
- `examples-content.tsx` — all logic: list, search, form dialog, delete confirm

**Use for:** CRUD pages with simple fields.
```

## When to Add an Example

- A new pattern is used repeatedly in the codebase
- A complex scenario is not covered by existing examples
- A new skill is created and has no real code examples yet
- An old pattern has changed and needs updating

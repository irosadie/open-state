---
name: web-slicing
description: Convert screenshot/Figma designs into pixel-perfect frontend UI with reusable components. Use when the task focuses on visual and layout implementation.
---

# Skill: Web Slicing

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

Implement UI from designs with precision. Focus on visuals and layout — no API integration yet.

## Workflow

1. Study the design: screenshot, Figma link, or layout description provided by the user.

2. Read the relevant guide before creating files in that folder:
   - `.agents/guides/web-page.md` — for files in `app/`
   - `.agents/guides/web-component.md` — for files in `components/`

3. Create files following the page structure convention:
   ```
   app/(group)/[feature]/
   ├── page.tsx               → Suspense wrapper (Server Component)
   └── [feature]-content.tsx → main Client Component (all logic: state, form, table, dialog)
   ```

4. Reference real code examples from:
   - `.agents/examples/web-slicing/nextjs-app-router/examples/` — basic CRUD template

5. Use components from the existing library:
   - `Button`, `Input`, `Dialog`, `Table`, `Select`, `RadioGroup`, `Textarea`
   - `ActionsDropdown` for edit/delete actions
   - `PanelCard` as table wrapper
   - `StatusBadge` for status display
   - `LoadingSpinner` for loading state
   - Lucide React for icons
   - Tailwind CSS for styling

6. Data not yet available from API → use dummy data / placeholder props.

## Prohibitions

- **FORBIDDEN** to call `axios`/`fetch` directly inside JSX/TSX components.
- **RECOMMENDED** to create `_components/` for components private to the route. Keep `*-page-content.tsx` beside `page.tsx` as the route orchestrator.
- **FORBIDDEN** to use `any` as a type.
- **FORBIDDEN** to hardcode colors or spacing — use Tailwind tokens.

## Pre-Completion Checklist

- [ ] `page.tsx` is a thin Suspense wrapper
- [ ] `*-page-content.tsx` sits beside `page.tsx` and orchestrates route state, hooks, and layout
- [ ] Route-specific components are in `_components/`; shared components are in `apps/web/components/`
- [ ] No direct API call in components
- [ ] No `any`
- [ ] Every file ends with a newline (EOF)

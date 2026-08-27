---
name: skill-creator
description: Create a new skill from scratch with full structure — SKILL.md, references, templates, openai.yaml, Claude wrapper, and registration in manifest + AGENTS.md + CLAUDE.md.
---

# Skill: Skill Creator

## Context (Required)
- Folder scope + conventions: `references/context.md`
- Execution checklist: `templates/checklist.md`

Create a complete, registered skill. `SKILL.md` remains the source of truth; `agents/openai.yaml` only holds Codex/OpenAI metadata such as `interface`, `policy`, or `dependencies`.

## Workflow

### 1. Define Skill Identity
- **Name** (kebab-case): `{scope}-{capability}` — e.g. `web-slicing`, `api-feature`, `ops-docker`
- **Scope**: `frontend`, `backend`, `docs`, `ops`, `flow`, or `meta`
- **One-sentence description**: when it applies and what it outputs
- **Trigger**: condition/keyword that activates this skill

### 1a. Use the Generator When Possible

For a compliant scaffold from the start, prefer the generator:

```bash
bun run skills:create -- \
  --name {scope}-{capability} \
  --scope {scope} \
  --description "{One-sentence description}" \
  --when "{When used in Skill Registry}"
```

The generator creates the skill folder, `openai.yaml` metadata, Claude wrapper, updates `manifest.json`, and adds registry entries.

### 2. Create Skill Folder & Files

Required structure in `.agents/skills/{skill-name}/`:

```
.agents/skills/{skill-name}/
├── SKILL.md                → Skill definition (name, description, workflow, prohibitions, checklist)
├── references/
│   └── context.md          → Target folder, real code examples, key patterns
├── templates/
│   └── checklist.md        → Step-by-step execution checklist
└── agents/
    └── openai.yaml         → Codex/OpenAI metadata (interface, policy, dependencies)
```

#### `SKILL.md` Format
```markdown
---
name: {skill-name}
description: {One-sentence description}
---

# Skill: {Title}

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

{One paragraph explaining the skill's purpose}

## Workflow

1. ...
2. ...

## Prohibitions

- **NEVER** ...

## Pre-Completion Checklist

- [ ] ...
```

#### `references/context.md` Format
```markdown
# Context: {Skill Title}

## Target Folder

\`\`\`
{folder tree that will be touched}
\`\`\`

## Real Code Examples

See: `.agents/examples/{scope}/{subfolder}/`

## Key Patterns

{Inline code examples for the main patterns}
```

#### `templates/checklist.md` Format
```markdown
# Checklist: {Skill Title}

- [ ] Read `.agents/settings.json`
- [ ] Read `.agents/guides/ARCHITECTURE.md`
- [ ] Read `references/context.md`
- [ ] ...skill-specific steps...
- [ ] All files end with a newline (EOF)
```

#### `agents/openai.yaml` Format
```yaml
interface:
  display_name: "{Short Skill Title}"
  short_description: "{Short description for the skill picker}"
  default_prompt: "Use $skill-name to {main outcome of this skill}."
policy:
  allow_implicit_invocation: true
```

### 3. Create Claude Wrapper

Create the file at `.claude/skills/{skill-name}/SKILL.md`:

```markdown
---
name: "{skill-name}"
description: "{One-sentence description}"
---

Source of truth lives at `.agents/skills/{skill-name}/SKILL.md`.

When this skill is used:
1. Read `.agents/skills/{skill-name}/SKILL.md`.
2. Follow the workflow and rules in that file.
3. Read referenced child files (`references/context.md`, `templates/checklist.md`) from the source-of-truth folder.
```

### 4. Register in `manifest.json`

Add an entry to `.agents/skills/manifest.json`:

```json
{
  "name": "{skill-name}",
  "description": "{One-sentence description}",
  "scope": "{scope}",
  "path": ".agents/skills/{skill-name}"
}
```

### 5. Update `AGENTS.md`

In `.agents/AGENTS.md`, add:
- A row to the **Skill Registry** table (name, scope, when used)
- A reference link in the **SKILL** section (when to read its SKILL.md)

### 6. Update `CLAUDE.md`

In the root `CLAUDE.md`:
- Add a row to the **Skill Registry** table
- Add a trigger entry in system-reminder (if Claude SDK is used)

### 7. Sync Claude Wrapper

After creating or editing a skill, run:

```bash
bun run skills:sync
```

### 8. Validate Skill Assets

Before considering the task done, run:

```bash
bun run skills:validate
```

## Prohibitions

- **NEVER** skip the Claude wrapper — always create `.claude/skills/{name}/SKILL.md`.
- **NEVER** create a skill without registering it in `manifest.json`.
- **NEVER** use an ambiguous skill name — must be `{scope}-{capability}`.
- **NEVER** duplicate an existing skill — check the manifest first.
- **NEVER** place main skill instructions in `agents/openai.yaml` — main instructions stay in `SKILL.md`.

## Pre-Completion Checklist

- [ ] Skill name defined (kebab-case, `{scope}-{capability}`)
- [ ] `SKILL.md` created with frontmatter + workflow + prohibitions + checklist
- [ ] `references/context.md` created with target folder + patterns
- [ ] `templates/checklist.md` created with full steps
- [ ] `agents/openai.yaml` created
- [ ] `.claude/skills/{name}/SKILL.md` created (Claude wrapper)
- [ ] Entry added in `manifest.json`
- [ ] `manifest.json` contains `whenToUse` for the registry
- [ ] Skill Registry in `.agents/AGENTS.md` and `CLAUDE.md` synced from script
- [ ] `bun run skills:sync` executed
- [ ] `bun run skills:validate` passes
- [ ] All files end with a newline (EOF)

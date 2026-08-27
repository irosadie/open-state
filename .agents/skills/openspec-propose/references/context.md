# Context: openspec-propose

## Target Folders
```
openspec/
└── changes/      → active changes
    └── archive/  → archived changes
openspec/specs/   → main specs
```

## CLI Reference
```bash
openspec list --json
openspec new change "<name>"
openspec status --change "<name>" --json
openspec instructions <artifact-id> --change "<name>" --json
openspec instructions apply --change "<name>" --json
```

Full instructions: `.claude/skills/openspec-propose/SKILL.md`

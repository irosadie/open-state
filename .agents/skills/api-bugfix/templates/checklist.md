# Checklist: API Bugfix

## Preparation

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Write down the bug symptom or failing behavior

## Execution

- [ ] Root cause localized to the smallest layer
- [ ] Changes remain minimal touch
- [ ] DTO/type/docs/OpenAPI updated when affected
- [ ] sqlc regenerated if query changed (`sqlc generate`)
- [ ] No unrelated refactor
- [ ] Reproduction or guard test added/updated

## Finalization

- [ ] Backend contract stays in sync
- [ ] OpenAPI updated if contract changed
- [ ] `go vet ./...` passes on touched surface
- [ ] `go build ./...` passes
- [ ] No unrelated changes carried along
- [ ] All files end with a newline (EOF)

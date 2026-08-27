# Checklist: docs-openapi

## Preparation

- [ ] Check OpenSpec specs/design under `openspec/changes/{slug}/` for endpoint requirements
- [ ] Check `docs/openapi/components/` — are the required schemas/responses already present?

## Path Files

- [ ] Create `docs/openapi/paths/{feature-slug}.yaml`
- [ ] All endpoints of the feature are covered
- [ ] Parameters (path, query, body) fully defined
- [ ] Success responses use the correct schema
- [ ] Error responses use shared `$ref` from components/responses/

## Schema Components

- [ ] Create a new entity schema in `components/schemas/` if missing
- [ ] No duplication — use `$ref` when a schema already exists
- [ ] Every field has the correct type (string, integer, boolean, etc.)
- [ ] `required` array lists the mandatory fields

## Entry File

- [ ] `docs/openapi/openapi.yaml` updated with `$ref` to the new path
- [ ] New schemas registered under `components.schemas`
- [ ] File parses cleanly in Swagger/Redoc

## Finalization

- [ ] Every file ends with a newline (EOF)
- [ ] No remaining placeholders

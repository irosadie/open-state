# Context: docs-openapi

## Existing Shared Components

### components/schemas/PaginationMeta.yaml
```yaml
type: object
properties:
  pagination:
    type: object
    properties:
      total:
        type: integer
      currentPage:
        type: integer
      perPage:
        type: integer
      lastPage:
        type: integer
  cursor:
    type: string
    nullable: true
```

### components/responses/NotFound.yaml
```yaml
description: Resource not found
content:
  application/json:
    schema:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
              example: NOT_FOUND
            message:
              type: string
```

### components/responses/Validation.yaml
```yaml
description: Validation failed
content:
  application/json:
    schema:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
              example: VALIDATION
            message:
              type: string
            fields:
              type: object
              additionalProperties:
                type: array
                items:
                  type: string
```

## Entity Schema Example

### components/schemas/PaymentMethod.yaml
```yaml
type: object
properties:
  id:
    type: string
    format: uuid
  name:
    type: string
  code:
    type: string
  isActive:
    type: boolean
  createdAt:
    type: string
    format: date-time
  updatedAt:
    type: string
    format: date-time
required: [id, name, code, isActive, createdAt, updatedAt]
```

## Adding a New Feature

1. Create `docs/openapi/paths/{feature-slug}.yaml`
2. Create `docs/openapi/components/schemas/{EntityName}.yaml` if it does not exist
3. Add `$ref` entries in `docs/openapi/openapi.yaml`
4. Reuse existing shared components — never duplicate

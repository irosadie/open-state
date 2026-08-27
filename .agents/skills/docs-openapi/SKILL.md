---
name: docs-openapi
description: Write or update OpenAPI documentation in split-per-feature format under `docs/openapi`. Use when endpoints, request/response schemas, or API contracts change.
---

# Skill: Docs OpenAPI

## Context (Required)
- Output folder: `docs/openapi/`
- Entry file: `docs/openapi/openapi.yaml`
- Per-feature split: `docs/openapi/paths/{feature-slug}.yaml`

Use this skill to translate API contracts into consistent, reusable, tool-parseable OpenAPI specs.

## Workflow

1. Read the API contract or finalized endpoints to document.
2. Determine whether the required shared schemas and responses already exist in `components/`.
3. Write or update `paths/{feature}.yaml` for the feature's endpoints.
4. Add the required `$ref` entries in `openapi.yaml`.
5. Validate that every `$ref` points to an existing file and key.

## File Structure

```
docs/openapi/
├── openapi.yaml              ← entry, $ref to paths + components
├── paths/
│   ├── payment-methods.yaml
│   └── users.yaml
└── components/
    ├── schemas/
    │   ├── PaymentMethod.yaml
    │   └── PaginationMeta.yaml
    └── responses/
        ├── NotFound.yaml
        └── Validation.yaml
```

## openapi.yaml Format (Entry)

```yaml
openapi: "3.0.3"
info:
  title: API
  version: "1.0.0"
paths:
  /payment-methods:
    $ref: "./paths/payment-methods.yaml#/PaymentMethodsList"
  /payment-methods/{id}:
    $ref: "./paths/payment-methods.yaml#/PaymentMethodsById"
components:
  schemas:
    PaymentMethod:
      $ref: "./components/schemas/PaymentMethod.yaml"
    PaginationMeta:
      $ref: "./components/schemas/PaginationMeta.yaml"
  responses:
    NotFound:
      $ref: "./components/responses/NotFound.yaml"
```

## paths/{feature}.yaml Format

```yaml
PaymentMethodsList:
  get:
    summary: List payment methods
    tags: [PaymentMethods]
    parameters:
      - name: page
        in: query
        schema:
          type: integer
          default: 1
      - name: perPage
        in: query
        schema:
          type: integer
          default: 10
    responses:
      "200":
        description: OK
        content:
          application/json:
            schema:
              type: object
              properties:
                list:
                  type: array
                  items:
                    $ref: "../components/schemas/PaymentMethod.yaml"
                meta:
                  $ref: "../components/schemas/PaginationMeta.yaml"
      "401":
        $ref: "../components/responses/Unauthorized.yaml"

  post:
    summary: Create payment method
    tags: [PaymentMethods]
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required: [name, code]
            properties:
              name:
                type: string
              code:
                type: string
              isActive:
                type: boolean
                default: true
    responses:
      "200":
        description: OK
        content:
          application/json:
            schema:
              type: object
              properties:
                data:
                  $ref: "../components/schemas/PaymentMethod.yaml"
      "409":
        $ref: "../components/responses/Conflict.yaml"
      "422":
        $ref: "../components/responses/Validation.yaml"
```

## Rules

- Use OpenAPI 3.0.3
- Split per feature — never dump all endpoints into one file
- `$ref` for any schema or response used more than once
- Tag = feature name (PascalCase)
- Every error response uses the shared `components/responses/`

## Prohibitions

- **NEVER** duplicate schemas that can be referenced from `components/`.
- **NEVER** place all repo endpoints in a single path file.
- **NEVER** leave a response body without an explicit schema.
- **NEVER** leave broken `$ref` entries or invalid placeholders.

## Pre-Completion Checklist

- [ ] `paths/{feature}.yaml` covers every endpoint of the feature
- [ ] `openapi.yaml` references the new path
- [ ] Shared schemas and responses are reused via `$ref`
- [ ] No placeholders or broken `$ref` entries
- [ ] Spec parses with standard OpenAPI tooling
- [ ] Every file ends with a newline (EOF)

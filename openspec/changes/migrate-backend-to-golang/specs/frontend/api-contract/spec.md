## MODIFIED Requirements

### Requirement: Frontend Zod schemas are frontend-only
`packages/schemas` Zod schemas SHALL be used exclusively by `apps/web`. The backend (Go) MUST NOT import from `packages/schemas`. Backend validation is handled by Go structs and inline validation logic.

#### Scenario: Frontend form validation uses packages/schemas
- **WHEN** a frontend form is submitted
- **THEN** validation SHALL use Zod schemas from `packages/schemas`

#### Scenario: Backend does not import packages/schemas
- **WHEN** the Go backend is compiled
- **THEN** it SHALL have no dependency on `packages/schemas` or any TypeScript package

### Requirement: API response types remain in packages/types
`packages/types` TypeScript types SHALL remain the source of truth for API response shapes consumed by `apps/web`. These types MUST stay in sync with Go response structs.

#### Scenario: Frontend uses packages/types for API response typing
- **WHEN** a react-query hook receives an API response
- **THEN** it SHALL be typed using types from `packages/types`

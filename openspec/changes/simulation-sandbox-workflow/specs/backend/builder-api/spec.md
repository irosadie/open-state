## ADDED Requirements

### Requirement: Simulate a workflow through the Builder API
The authenticated Builder API SHALL expose `POST /api/workflows/simulate` for
simulating a supplied workflow snapshot. The request SHALL contain the workflow
definition and MAY contain initial context and an ordered event script. The response
SHALL be wrapped in `{ "data": ... }` and contain the structured sandbox trace.

The tenant SHALL be derived from `X-Tenant-ID`, never from the request body. The
endpoint SHALL accept an unsaved workflow snapshot and SHALL not require a workflow
id or published version.

#### Scenario: Authenticated operator simulates an unsaved draft
- **WHEN** an authenticated request with `X-Tenant-ID` posts a valid draft snapshot
  to `/api/workflows/simulate`
- **THEN** the API returns the draft's sandbox trace in a successful `{ "data": ... }`
  response

#### Scenario: Request omits the tenant header
- **WHEN** a simulation request does not include `X-Tenant-ID`
- **THEN** the API rejects it using the same tenant-authentication behavior as the
  other Builder API operations

#### Scenario: Invalid simulation payload
- **WHEN** a simulation request lacks a workflow definition or has malformed event
  input
- **THEN** the API returns a structured validation error and no server state changes

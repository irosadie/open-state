## Purpose

Defines the DomainError type and its mapping to HTTP status codes in the Go backend, enabling consistent error responses across all handlers.

## ADDED Requirements

### Requirement: DomainError is a typed Go error
The system SHALL define a `DomainError` struct implementing the `error` interface with a `Code` field (string) and `Message` field (string).

#### Scenario: DomainError satisfies error interface
- **WHEN** a `DomainError` is returned from a use case
- **THEN** it SHALL be assignable to the standard Go `error` interface

### Requirement: Error codes map to HTTP status codes
The system SHALL map the following DomainError codes to HTTP status codes:

| Code | HTTP Status |
|---|---|
| NOT_FOUND | 404 |
| UNAUTHORIZED | 401 |
| FORBIDDEN | 403 |
| CONFLICT | 409 |
| VALIDATION | 422 |
| INTERNAL | 500 |

#### Scenario: Correct HTTP status returned for each code
- **WHEN** a handler returns a DomainError with a known code
- **THEN** the Echo error handler SHALL respond with the corresponding HTTP status and a JSON body `{"error": "<message>"}`

### Requirement: Unknown errors default to 500
The system SHALL treat any error that is not a DomainError as an INTERNAL error.

#### Scenario: Non-DomainError returns 500
- **WHEN** a handler returns a plain Go error (not DomainError)
- **THEN** the response SHALL have HTTP status 500 and a generic error message

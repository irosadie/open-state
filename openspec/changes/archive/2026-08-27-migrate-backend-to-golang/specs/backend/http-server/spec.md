## Purpose

Provides the HTTP server entrypoint for the backend API, including middleware chain, routing, and centralized error handling using Echo framework.

## ADDED Requirements

### Requirement: Server starts and binds to configured port
The system SHALL start an HTTP server bound to the port defined in environment configuration.

#### Scenario: Successful server start
- **WHEN** the server process is started with a valid PORT env var
- **THEN** the server listens on that port and accepts HTTP connections

#### Scenario: Missing port configuration
- **WHEN** PORT env var is not set
- **THEN** the server SHALL default to port 8080

### Requirement: Middleware chain is applied globally
The system SHALL apply the following middleware to all routes in order: request logger, CORS, request ID, and authentication session validation.

#### Scenario: Request passes through middleware
- **WHEN** an HTTP request is received
- **THEN** it passes through logger → CORS → request-id → auth middleware before reaching the handler

#### Scenario: CORS headers present on response
- **WHEN** a request includes an Origin header
- **THEN** the response SHALL include appropriate CORS headers

### Requirement: Centralized error handler
The system SHALL map DomainErrors to HTTP status codes via a single global error handler registered on the Echo instance.

#### Scenario: DomainError NOT_FOUND mapped to 404
- **WHEN** a handler returns a DomainError with code NOT_FOUND
- **THEN** the response SHALL have HTTP status 404 with a JSON error body

#### Scenario: DomainError UNAUTHORIZED mapped to 401
- **WHEN** a handler returns a DomainError with code UNAUTHORIZED
- **THEN** the response SHALL have HTTP status 401

#### Scenario: DomainError CONFLICT mapped to 409
- **WHEN** a handler returns a DomainError with code CONFLICT
- **THEN** the response SHALL have HTTP status 409

#### Scenario: DomainError VALIDATION mapped to 422
- **WHEN** a handler returns a DomainError with code VALIDATION
- **THEN** the response SHALL have HTTP status 422

#### Scenario: Unhandled error mapped to 500
- **WHEN** a handler returns a non-DomainError
- **THEN** the response SHALL have HTTP status 500

### Requirement: Health endpoint available
The system SHALL expose a `GET /health` endpoint that returns server status without authentication.

#### Scenario: Health check returns 200
- **WHEN** `GET /health` is called
- **THEN** the response SHALL have HTTP status 200 with JSON body indicating status ok

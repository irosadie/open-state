## Purpose

Manages user registration, login, logout, and current-user retrieval with JWT-based authentication and server-side session tracking.

## ADDED Requirements

### Requirement: User registration
The system SHALL allow a new user to register with email, password, and name.

#### Scenario: Successful registration
- **WHEN** a POST request is made to `/api/auth/register` with valid email, password, and name
- **THEN** the system SHALL create the user, return HTTP 201 with user data (id, email, name, role)

#### Scenario: Duplicate email rejected
- **WHEN** a POST request is made to `/api/auth/register` with an email already in use
- **THEN** the system SHALL return a DomainError CONFLICT

#### Scenario: Invalid input rejected
- **WHEN** registration payload is missing required fields or email format is invalid
- **THEN** the system SHALL return a DomainError VALIDATION

### Requirement: User login
The system SHALL authenticate a user by email and password and return a JWT token.

#### Scenario: Successful login
- **WHEN** a POST request is made to `/api/auth/login` with valid credentials
- **THEN** the system SHALL return HTTP 200 with a signed JWT access token and session record created in DB

#### Scenario: Invalid credentials rejected
- **WHEN** login is attempted with wrong password or non-existent email
- **THEN** the system SHALL return a DomainError UNAUTHORIZED

### Requirement: User logout
The system SHALL invalidate the current session on logout.

#### Scenario: Successful logout
- **WHEN** an authenticated POST request is made to `/api/auth/logout`
- **THEN** the system SHALL delete the session record and return HTTP 200

#### Scenario: Unauthenticated logout rejected
- **WHEN** logout is called without a valid JWT
- **THEN** the system SHALL return a DomainError UNAUTHORIZED

### Requirement: Get current user
The system SHALL return the authenticated user's profile.

#### Scenario: Successful get current user
- **WHEN** an authenticated GET request is made to `/api/auth/me`
- **THEN** the system SHALL return HTTP 200 with user data (id, email, name, role, status, photo)

#### Scenario: Expired or invalid token rejected
- **WHEN** `GET /api/auth/me` is called with an expired or invalid JWT
- **THEN** the system SHALL return a DomainError UNAUTHORIZED

### Requirement: Password stored as hash
The system SHALL never store plaintext passwords. Passwords MUST be hashed using bcrypt before persistence.

#### Scenario: Password hash verified on login
- **WHEN** a user logs in
- **THEN** the system SHALL compare the provided password against the stored bcrypt hash

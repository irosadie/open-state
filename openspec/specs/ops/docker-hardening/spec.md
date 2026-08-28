# ops/docker-hardening Specification

## Purpose
Define a hardened Dockerfile for the API that builds a small, non-root,
minimal-distribution image for Linux deployment (PRD §84). It improves the
security posture by running as a non-root user, using multi-stage builds, and
minimizing the runtime base.

## Requirements

### Requirement: Multi-stage build

The platform SHALL build the API binary in a multi-stage Dockerfile.

- A builder stage SHALL compile the Go binary with CGO disabled and static
  linking where possible.
- A separate runtime stage SHALL contain only the binary and required
  certificates/timezone data.
- The build SHALL produce a minimal final image.

#### Scenario: Static binary build

- **WHEN** the image is built
- **THEN** the Go binary is built statically (CGO_ENABLED=0) so it runs on a
  minimal base

### Requirement: Non-root runtime user

The container SHALL run the process as a non-root user.

- A non-root user SHALL be created in the runtime stage (e.g. `nobody` or a
  dedicated `app` user).
- The `USER` directive SHALL switch to the non-root user before `CMD`.
- The binary and any required writable paths SHALL be owned by the non-root user.

#### Scenario: Process runs non-root

- **WHEN** the container starts
- **THEN** the API process runs with a non-root UID/GID

#### Scenario: Non-root filesystem permissions

- **WHEN** the process needs to write (e.g. temp/storage)
- **THEN** the non-root user has write access to only those paths

### Requirement: Minimal/distroless runtime base

The platform SHALL prefer a minimal runtime base (e.g. `scratch` or
`distroless/static`).

- A `distroless/static` or `scratch` base SHALL be used for the runtime stage.
- If a shell is required, a minimal Alpine base SHALL be acceptable, but the
  default SHALL be distroless/static.

#### Scenario: Distroless runtime

- **WHEN** the image is built with the default config
- **THEN** the runtime stage uses a distroless/static base

### Requirement: Read-only filesystem where possible

The container SHALL run with a read-only root filesystem where feasible.

- A read-only rootfs SHOULD be supported (document `--read-only` usage).
- Only explicitly writable volumes (e.g. upload storage) SHALL be writable.

#### Scenario: Read-only rootfs supported

- **WHEN** the container runs with `--read-only`
- **THEN** the API starts and only the configured writable volume is writable

### Requirement: No unnecessary packages

The runtime image SHALL NOT contain compilers, shells, or build tooling.

- Only runtime dependencies (CA certs, tzdata) SHALL be present.
- Secrets SHALL NOT be baked into the image (see `ops/secrets-management`).

#### Scenario: Minimal dependencies

- **WHEN** the runtime image is inspected
- **THEN** it contains no compilers/shells and only required runtime files

### Requirement: Healthcheck

The image SHALL declare a healthcheck against the API health endpoint.

- A `HEALTHCHECK` SHALL call `GET /health` (or an equivalent) on the API port.

#### Scenario: Container healthcheck

- **WHEN** the container runs
- **THEN** the healthcheck periodically verifies the health endpoint

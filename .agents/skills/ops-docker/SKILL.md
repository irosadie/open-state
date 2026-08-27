---
name: ops-docker
description: Write or modify Dockerfiles for this backend so it is ready to deploy on Linux. Use for tasks involving containerization, image optimization, or build/runtime issues in containers. Docker Compose is managed by the server — do not touch it.
---

# Skill: Ops Docker

## Context (Required)
- Target: Dockerfile in `apps/api/` and/or `apps/worker/`
- Stack: Go (Echo + sqlc for api, asynq for worker)
- **Do NOT touch `docker-compose.yml`** — managed by the server

## Principles

- Multi-stage build to minimize image size
- `builder` stage: compile Go binary
- `runner` stage: minimal alpine with binary only
- Run as non-root user
- Build context is monorepo root so `packages/go-shared` and `go.work` are available

## Workflow

1. Identify target app and its runtime needs.
2. Copy `go.work`, `go.work.sum`, `packages/go-shared/`, and the target app in builder stage.
3. Build Go binary in `builder` stage.
4. Copy only the binary into `runner` stage.
5. Verify `CMD`, port, and env var requirements before finishing.

## Dockerfile Template (Echo API)

```dockerfile
# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum* ./
COPY packages/go-shared/ ./packages/go-shared/
COPY apps/api/ ./apps/api/

WORKDIR /app/apps/api
RUN go build -o /bin/server ./cmd/server/main.go

# Stage 2: Runner
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /bin/server .

USER appuser
EXPOSE 8080
CMD ["./server"]
```

## Dockerfile Template (asynq Worker)

```dockerfile
# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum* ./
COPY packages/go-shared/ ./packages/go-shared/
COPY apps/worker/ ./apps/worker/

WORKDIR /app/apps/worker
RUN go build -o /bin/worker ./cmd/worker/main.go

# Stage 2: Runner
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /bin/worker .

USER appuser
CMD ["./worker"]
```

## Rules

- Always build from monorepo root so `go.work` resolves `packages/go-shared`
- Never copy `.env` into the image — inject via environment variable at runtime
- Never expose unused ports
- Copy `go.work.sum*` with glob to handle optional file gracefully

## Prohibitions

- **FORBIDDEN** to modify `docker-compose.yml`.
- **FORBIDDEN** to run container as root unless strictly required.
- **FORBIDDEN** to copy the entire repo into the runner stage.
- **FORBIDDEN** to leave a Dockerfile that cannot be built deterministically.

## Pre-Completion Checklist

- [ ] Dockerfile uses multi-stage build (golang:1.26-alpine builder, alpine:3.20 runner)
- [ ] `go.work` and `packages/go-shared/` copied in builder stage
- [ ] Runner stage contains only the compiled binary
- [ ] Non-root user is used
- [ ] Port and CMD match target app
- [ ] All files end with newline (EOF)

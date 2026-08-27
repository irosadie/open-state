# Checklist: ops-docker

## Preparation

- [ ] Identify target: `apps/api` and/or `apps/worker`
- [ ] Check ports exposed in source code (`cmd/server/main.go`)
- [ ] Confirm `go.work` and `packages/go-shared/` are at repo root

## Dockerfile

- [ ] Multi-stage build (builder + runner)
- [ ] Builder image: `golang:1.26-alpine`
- [ ] Runner image: `alpine:3.20`
- [ ] `go.work`, `go.work.sum*`, and `packages/go-shared/` copied in builder stage
- [ ] Go binary built with `go build -o /bin/<name> ./cmd/<name>/main.go`
- [ ] Only compiled binary copied into runner stage
- [ ] Non-root user created and used in runner stage
- [ ] `ca-certificates` and `tzdata` installed in runner stage
- [ ] `EXPOSE` matches used port (api: 8080)
- [ ] `CMD` runs compiled binary (`["./server"]` or `["./worker"]`)

## Security

- [ ] No `.env` or credentials in image
- [ ] Non-root user
- [ ] No build tools or source code in runner stage

## Validation

- [ ] Build context is monorepo root: `docker build -f apps/{app}/Dockerfile -t test:latest .`
- [ ] Container runs: `docker run --rm -e DATABASE_URL=... -e JWT_SECRET=... test:latest`
- [ ] Image size reasonable (Go Alpine binary typically < 50MB)

## Finalization

- [ ] Dockerfile ends with newline (EOF)
- [ ] `.dockerignore` exists at root if not present
- [ ] **Do NOT modify `docker-compose.yml`**

# Context: ops-docker

## Target Files

```
apps/api/Dockerfile
apps/worker/Dockerfile
```

## Monorepo Structure in Container (build context = root)

```
/app/
├── go.work
├── go.work.sum
├── packages/
│   └── go-shared/    ← shared Go module (DomainError)
└── apps/
    ├── api/          ← Echo + sqlc + goose
    └── worker/       ← asynq
```

## Environment Variables

Never hardcode in Dockerfile. Inject at `docker run` or via orchestrator:

```bash
docker run \
  -e DATABASE_URL="postgresql://..." \
  -e REDIS_URL="redis://..." \
  -e JWT_SECRET="..." \
  -p 8080:8080 \
  my-api:latest
```

## Build Command

```bash
# Build context MUST be monorepo root (for go.work + packages/go-shared)
docker build -f apps/api/Dockerfile -t my-api:latest .
docker build -f apps/worker/Dockerfile -t my-worker:latest .
```

## Layer Caching Tips

Optimal COPY order for cache hits:
1. `go.work` + `go.work.sum` (rarely change)
2. `packages/go-shared/` (rarely change)
3. `apps/{api,worker}/go.mod` + `go.sum` (changes when deps change)
4. Source code (changes often — place last)

## .dockerignore

```
node_modules
.env
.env.*
dist
.git
apps/web
docs
*.md
```

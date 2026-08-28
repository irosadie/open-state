# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go workspace files (build context is monorepo root)
COPY go.work go.work.sum* ./
COPY packages/go-shared/ ./packages/go-shared/
COPY apps/api/ ./apps/api/

# Build a static binary (CGO disabled) so it runs on a minimal base
WORKDIR /app/apps/api
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server/main.go

# Runtime stage: minimal alpine with only the binary + CA certs + tzdata.
# Runs as a non-root user; no compilers/shells from the builder are present.
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /bin/server .

USER appuser
EXPOSE 8020

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8020/health || exit 1

CMD ["./server"]

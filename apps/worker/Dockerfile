# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go workspace files
COPY go.work go.work.sum* ./
COPY packages/go-shared/ ./packages/go-shared/
COPY apps/worker/ ./apps/worker/

# Build
WORKDIR /app/apps/worker
RUN go build -o /bin/worker ./cmd/worker/main.go

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/worker .

CMD ["./worker"]

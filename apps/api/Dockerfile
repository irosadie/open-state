# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go workspace files
COPY go.work go.work.sum* ./
COPY packages/go-shared/ ./packages/go-shared/
COPY apps/api/ ./apps/api/

# Build
WORKDIR /app/apps/api
RUN go build -o /bin/server ./cmd/server/main.go

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/server .

EXPOSE 8020

CMD ["./server"]

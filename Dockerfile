FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies (git, gcc, musl-dev)
RUN apk add --no-cache git gcc musl-dev

# Copy only go.mod (go.sum is intentionally omitted)
COPY backend/go.mod backend/go.mod

# Download dependencies – no go.sum, so verification is skipped
RUN cd backend && go mod download

# Copy the rest of the source code
COPY backend/ backend/

# Build the application (CGO enabled for SQLite compatibility, though we use MySQL)
RUN cd backend && CGO_ENABLED=1 GOOS=linux go build -o /build/app .

# Final lightweight image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /build/app /app/app

# Copy frontend static files
COPY frontend/ /app/frontend/

# Create a non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app && \
    chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/app"]

FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy only go.mod (go.sum will be generated during build)
COPY backend/go.mod backend/go.mod

# Download dependencies and generate a correct go.sum
RUN cd backend && go mod download && go mod tidy

# Copy the rest of the source code
COPY backend/ backend/

# Build the application (CGO enabled for compatibility)
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

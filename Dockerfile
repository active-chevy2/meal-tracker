FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go files
COPY backend/go.mod backend/go.mod
COPY backend/go.sum backend/go.sum

RUN cd backend && go mod download

# Copy source code
COPY backend/ backend/

# Build the application
RUN cd backend && CGO_ENABLED=1 GOOS=linux go build -o /build/app .

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app /app/app

# Copy frontend files
COPY frontend/ /app/frontend/

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app && \
    chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/app"]

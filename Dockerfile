# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies including GCC for CGO/sqlite3
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# CGO is required for sqlite3
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o modbridge ./main.go

# Final stage - use alpine for health checks and minimal size
FROM alpine:3.23

LABEL org.opencontainers.image.title="ModBridge" \
      org.opencontainers.image.description="Modbus TCP Proxy Manager" \
      org.opencontainers.image.source="https://github.com/Xerolux/modbridge" \
      org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -u 1000 -g appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/modbridge .

# Create directory for logs and config with correct permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080
EXPOSE 5020-5030

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./modbridge"]

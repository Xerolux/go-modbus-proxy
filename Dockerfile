# Multi-stage build for optimized production image
# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=1.0.0 -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o modbridge ./cmd/server

# Build CLI tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o modbusctl ./cmd/modbusctl

# Stage 2: Runtime
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 modbridge && \
    adduser -D -u 1000 -G modbridge modbridge

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/modbridge /app/modbridge
COPY --from=builder /build/modbusctl /usr/local/bin/modbusctl

# Copy default config
COPY config.json.example /app/config.json

# Create data directory
RUN mkdir -p /data && chown -R modbridge:modbridge /app /data

# Switch to non-root user
USER modbridge

# Expose ports
# 8080 - HTTP API
# 5020-5029 - Modbus proxy ports
EXPOSE 8080 5020 5021 5022 5023 5024 5025 5026 5027 5028 5029

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/modbusctl", "metrics", "health", "--api-url", "http://localhost:8080"]

# Set environment variables
ENV MODBRIDGE_WEB_PORT=:8080
ENV MODBRIDGE_LOG_LEVEL=info

# Volume for persistent data
VOLUME ["/data"]

# Run the application
ENTRYPOINT ["/app/modbridge"]
CMD ["-config", "/app/config.json"]

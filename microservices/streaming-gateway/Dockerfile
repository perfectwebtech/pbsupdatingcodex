# ============================================================================
# Multi-stage Dockerfile for IPTV Streaming Gateway
# ============================================================================

# ============================================================================
# Stage 1: Builder
# ============================================================================
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o streaming-gateway ./cmd/server

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/streaming-gateway .

# Copy configuration files
COPY --from=builder /build/configs ./configs

# Create necessary directories
RUN mkdir -p /app/logs /app/uploads /app/transcoded /app/tmp && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["./streaming-gateway"]

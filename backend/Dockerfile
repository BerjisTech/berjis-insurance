# Insurance Broker Platform - Backend Dockerfile
# Multi-stage build for optimized production image
# Go 1.22+ with Fiber framework

# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 for SQLite support if needed
# -ldflags for smaller binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o insurance-api \
    ./cmd/api

# =============================================================================
# Stage 2: Runtime
# =============================================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S insurance && \
    adduser -u 1001 -S insurance -G insurance

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/insurance-api .

# Create directories for uploads and logs
RUN mkdir -p /app/uploads /app/logs && \
    chown -R insurance:insurance /app

# Switch to non-root user
USER insurance

# Expose application port
EXPOSE 8096

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
    CMD wget --quiet --tries=1 --spider http://localhost:8096/health || exit 1

# Run the application
ENTRYPOINT ["/app/insurance-api"]

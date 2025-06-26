# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build-linux

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S fern && \
    adduser -u 1001 -S fern -G fern

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/fern-platform-linux ./fern-platform

# Copy configuration files
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/web ./web

# Change ownership
RUN chown -R fern:fern /app

# Switch to non-root user
USER fern

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./fern-platform", "-config", "config/config.yaml"]
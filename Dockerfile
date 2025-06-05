# Build stage
FROM golang:1.23.8-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod ./
# go.sum might not exist if there are no dependencies
COPY go.su[m] ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o p2p ./cmd/p2p

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 p2p && \
    adduser -D -u 1001 -G p2p p2p

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/p2p .

# Copy configuration files
COPY --from=builder /app/example-config.yaml ./config.yaml

# Change ownership to non-root user
RUN chown -R p2p:p2p /app

# Switch to non-root user
USER p2p

# Expose default port (will be overridden by config)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD netstat -an | grep :8080 > /dev/null; if [ 0 != $? ]; then exit 1; fi;

# Default command
CMD ["./p2p", "-config", "config.yaml"]
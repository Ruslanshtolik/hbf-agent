# Multi-stage build for HBF Agent

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev linux-headers

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o hbf-agent \
    ./cmd/hbf-agent

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    iptables \
    ip6tables \
    ca-certificates \
    tzdata

# Create directories
RUN mkdir -p /etc/hbf-agent /var/log/hbf-agent

# Copy binary from builder
COPY --from=builder /build/hbf-agent /usr/local/bin/hbf-agent

# Copy default configuration
COPY config/config.example.yaml /etc/hbf-agent/config.yaml

# Set permissions
RUN chmod +x /usr/local/bin/hbf-agent

# Expose ports
EXPOSE 9090 9091 9092 8080 8081

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/hbf-agent"]
CMD ["--config", "/etc/hbf-agent/config.yaml"]

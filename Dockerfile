# Multi-stage build for Go URL Shortener
# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o url-shortener .

# Stage 2: Runtime stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/url-shortener .

# Expose the port (configurable via environment)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./url-shortener"]

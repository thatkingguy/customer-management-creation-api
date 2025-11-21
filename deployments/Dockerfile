# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -a -installsuffix cgo \
    -o customer-service \
    ./cmd/server

# Runtime stage - using distroless for minimal size and security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy the binary from builder
COPY --from=builder /app/customer-service /customer-service

# Expose port
EXPOSE 8080

# Use nonroot user (distroless provides this automatically)
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["/customer-service"]


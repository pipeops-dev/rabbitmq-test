# Stage 1: Build the Go app
FROM golang:1.24.1-alpine AS builder

# Set necessary Go environment variables
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install git (required for go get in some cases)
RUN apk add --no-cache git

# Create working directory
WORKDIR /app

# Copy go.mod and go.sum first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Stage 2: Minimal image with compiled binary
FROM scratch

# Copy binary from builder
COPY --from=builder /app/main /main

# Set binary as entrypoint
ENTRYPOINT ["/main"]

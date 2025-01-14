# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /go/src/fowergram

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /go/bin/app ./cmd/api

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /go/bin/app .

# Create directory for source code
RUN mkdir -p /app/src

# Expose port
EXPOSE 8080

# Run the application
CMD ["./app"]
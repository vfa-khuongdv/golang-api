# Build stage: Use an official Go runtime as a parent image
FROM golang:1.25.5-alpine AS builder

# Set environment variables for Go build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download all the dependencies. Dependencies will be cached if go.mod and go.sum are unchanged
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go app with optimizations
RUN go build -ldflags="-s -w" -o main ./cmd/server/main.go

# Runtime stage: Use minimal image
FROM alpine:3.21

# Install ca-certificates for SSL (if required)
RUN apk --no-cache add ca-certificates

# Add a new non-root user
RUN adduser -D -u 1000 appuser

# Set the Current Working Directory inside the container
WORKDIR /home/appuser/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/main .

# Copy the internal directory which includes migrations
COPY --from=builder /app/internal ./internal

# Change ownership of the application files to the non-root user
RUN chown -R appuser:appuser /home/appuser

# Switch to the non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 3000

# Run the Go binary
CMD ["./main"]

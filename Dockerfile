FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO (SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum first to leverage cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the executable
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o albook .

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install explicit runtime dependencies if any (none for static binary usually, but sqlite libs might be needed if dynamic linking)
# With CGO on Alpine, usually it is dynamically linked against musl. 
# We'll install minimal packages just in case, or verify if it runs.
# Alpine's musl is compatible.

# Copy binary from builder
COPY --from=builder /app/albook .

# Create volume directory for database
RUN mkdir -p /data
VOLUME /data

# Expose port
EXPOSE 2100

# Run
CMD ["./albook", "-port", "2100", "-db", "/data/albook.db"]

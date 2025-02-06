# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o asn-labels

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ttf-dejavu

# Copy the binary from builder
COPY --from=builder /app/asn-labels .

# Create a directory for temporary files
RUN mkdir -p /tmp/asn-labels && \
    chmod 777 /tmp/asn-labels

# Expose the HTTP port
EXPOSE 8080

# Run the server
ENTRYPOINT ["./asn-labels", "-serve"]
CMD ["-port", "8080"]
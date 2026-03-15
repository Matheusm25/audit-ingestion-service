# Build stage
FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o audit-ingestion-service ./cmd/audit-ingestion-service

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/audit-ingestion-service .

# Expose HTTP port
EXPOSE 80

# Run the application
CMD ["./audit-ingestion-service"]

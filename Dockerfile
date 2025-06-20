# Multi-stage build for IoT Sensor Fleet applications

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the applications
RUN make build

# Final stage for sensor-producer
FROM alpine:3.18 AS sensor-producer

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/sensor-producer .

# Expose metrics port
EXPOSE 2112

# Command to run the application
CMD ["./sensor-producer"]

# Final stage for anomaly-detector
FROM alpine:3.18 AS anomaly-detector

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/anomaly-detector .

# Expose metrics port
EXPOSE 2113

# Command to run the application
CMD ["./anomaly-detector"]

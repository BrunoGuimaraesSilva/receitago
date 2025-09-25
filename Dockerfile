# ---------- Build stage ----------
FROM golang:1.25-alpine AS builder

# Install git & build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (for dependency caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the app
RUN go build -o receitago ./cmd/api

# ---------- Run stage ----------
FROM alpine:3.18

# Install certificates (needed for HTTPS requests)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/receitago .

# Create data directory (for downloads)
RUN mkdir -p /app/data

# Expose API port
EXPOSE 8080

# Run the service
CMD ["./receitago"]

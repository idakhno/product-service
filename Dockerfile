# syntax=docker/dockerfile:1

# Stage 1: Install migration tool
FROM golang:1.24-alpine AS migrate
RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Stage 2: Build application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /product-api ./cmd/api

# Stage 3: Final image
FROM alpine:3.20 AS final

# Copy migration tool
COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/

# Create non-privileged user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary and migrations (before switching user, as COPY requires root)
COPY --from=builder /product-api /product-api
COPY migrations ./migrations

# Switch to non-privileged user
USER appuser

# Expose port for HTTP server
EXPOSE 8080

# Run application
CMD ["/product-api"]
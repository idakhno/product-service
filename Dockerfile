# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS migrate
RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /product-api ./cmd/api

FROM alpine:3.20 AS final

COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

COPY --from=builder /product-api /product-api
COPY migrations ./migrations

EXPOSE 8080

CMD ["/product-api"]
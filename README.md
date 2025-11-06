# Product API

RESTful API service for managing users, products, and orders.

## Description

Product API is a microservice providing REST API for working with users, products, and orders. The service is implemented in Go using PostgreSQL as the database.

## Technology Stack

- **Go 1.24** - Programming language
- **PostgreSQL 15** - Relational database
- **Chi** - HTTP router
- **Docker & Docker Compose** - Containerization
- **OpenTelemetry** - Distributed tracing
- **Sentry** - Error monitoring
- **Swagger** - API documentation

## Quick Start

### Requirements

- Docker and Docker Compose
- Make (optional, for convenience)

### Installation and Running

1. **Clone the repository**

```bash
git clone <repository-url>
cd product-service
```

2. **Start services**

```bash
make up
# or
docker-compose up -d --build
```

3. **Apply migrations**

```bash
make migrate-up
# or
docker-compose run --rm api migrate -path /app/migrations -database ${DATABASE_URL} -verbose up
```

4. **Verify API is working**

API will be available at: `http://localhost:8080`

Swagger documentation: `http://localhost:8080/swagger/index.html`

## API Usage

### User Registration

```bash
curl -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "firstname": "John",
    "lastname": "Doe",
    "age": 30,
    "is_married": false
  }'
```

### Authentication

```bash
curl -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

The response contains a JWT token that should be used in the `Authorization: Bearer <token>` header for protected endpoints.

### Create Product

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "description": "High-quality wireless headphones",
    "tags": ["audio", "electronics", "wireless"],
    "quantity": 100,
    "price": 99.99
  }'
```

### Create Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "items": [
      {
        "product_id": "product-uuid-here",
        "quantity": 2
      }
    ]
  }'
```

## Available Commands

### Make Commands

- `make help` - Show help for all commands
- `make up` - Start all services
- `make down` - Stop all services
- `make logs` - Show service logs
- `make migrate-up` - Apply migrations
- `make migrate-down` - Rollback migrations
- `make test` - Run tests
- `make lint` - Run linter
- `make swagger` - Generate Swagger documentation
- `make build` - Build binary file
- `make run` - Run application locally (without Docker)
- `make install-tools` - Install development tools

## Development

### Installing Development Tools

```bash
make install-tools
```

### Local Run (without Docker)

1. Ensure PostgreSQL is running and accessible
2. Configure environment variables (DATABASE_URL, JWT_SECRET, etc.)
3. Run the application:

```bash
make run
```

### Running Tests

```bash
make test
```

### Code Linting

```bash
make lint
```

### Generating Swagger Documentation

```bash
make swagger
```

## Security

- Passwords are hashed using bcrypt
- JWT tokens are used for authentication
- Protected endpoints require a valid JWT token
- Application runs as non-privileged user in Docker

## Monitoring

- **Sentry** - Error and exception tracking
- **OpenTelemetry** - Distributed request tracing
- **Structured logging** - Structured logging using slog

## License

MIT

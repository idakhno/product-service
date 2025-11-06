.PHONY: run test lint swagger compose-up compose-down compose-logs migrate-up migrate-down install-tools mockery build clean help

# Binary file name
BINARY_NAME=product-api

# Development tool paths
GOLANGCILINT_BIN=golangci-lint
MIGRATE_BIN=migrate
SWAG_BIN=swag

.DEFAULT_GOAL := help

# help: Shows help for all available commands
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# install-tools: Installs all necessary development tools
install-tools: ## Install all necessary Go development tools
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/vektra/mockery/v2@v2.42.1

# run: Runs application locally (without Docker)
run: ## Run the application (locally, without Docker)
	@go run ./cmd/api

# test: Runs unit and integration tests (requires running database)
test: ## Run unit and integration tests (requires 'db' service to be running)
	@echo "Ensure database is running: make compose-up"
	@go test -v ./...

# lint: Runs linter for code checking
lint: ## Run linter
	@if command -v $(GOLANGCILINT_BIN) > /dev/null; then \
		$(GOLANGCILINT_BIN) run ./...; \
	else \
		echo "golangci-lint not found. Install it with: make install-tools"; \
		exit 1; \
	fi

# swagger: Generates Swagger documentation
swagger: ## Generate swagger docs
	@if command -v $(SWAG_BIN) > /dev/null; then \
		$(SWAG_BIN) init -g cmd/api/main.go; \
	else \
		echo "swag not found. Install it with: make install-tools"; \
		exit 1; \
	fi

# mockery: Generates mocks for interfaces
mockery: ## Generate mocks
	@if command -v mockery > /dev/null; then \
		mockery --all --case=underscore --with-expecter; \
	else \
		echo "mockery not found. Install it with: make install-tools"; \
		exit 1; \
	fi

# build: Builds application binary
build: ## Build the application binary
	@go build -o bin/$(BINARY_NAME) ./cmd/api

# clean: Removes compiled files
clean: ## Clean build artifacts
	@rm -rf bin/

# up: Builds and starts API and database in background mode
up: ## Build and start API and database services in detached mode
	@docker-compose up -d --build

# down: Stops and removes all services
down: ## Stop and remove all services
	@docker-compose down

# logs: Shows logs of all services
logs: ## Follow logs of all services
	@docker-compose logs -f

# migrate-up: Applies database migrations
migrate-up: ## Apply database migrations
	@docker-compose run --rm api migrate -path /app/migrations -database ${DATABASE_URL} -verbose up

# migrate-down: Rolls back database migrations
migrate-down: ## Rollback database migrations
	@docker-compose run --rm api migrate -path /app/migrations -database ${DATABASE_URL} -verbose down
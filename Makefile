.PHONY: run test lint swagger compose-up compose-down compose-logs migrate-up migrate-down install-tools mockery

BINARY_NAME=product-api

GOLANGCILINT_BIN=golangci-lint
MIGRATE_BIN=migrate
SWAG_BIN=swag

.DEFAULT_GOAL := help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

install-tools: ## Install all necessary Go development tools
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/vektra/mockery/v2@v2.42.1

run: ## Run the application (locally, without Docker)
	@go run ./cmd/api

test: ## Run unit and integration tests (requires 'db' service to be running)
	@echo "Ensure database is running: make compose-up"
	@go test -v ./...

lint: ## Run linter
	@$(GOLANGCILINT_BIN) run ./...

swagger: ## Generate swagger docs
	@$(SWAG_BIN) init -g cmd/api/main.go

mockery: ## Generate mocks
	@mockery --all --case=underscore --with-expecter

up: ## Build and start API and database services in detached mode
	@docker-compose up -d --build

down: ## Stop and remove all services
	@docker-compose down

logs: ## Follow logs of all services
	@docker-compose logs -f

migrate-up: ## Apply database migrations
	@docker-compose run --rm api migrate -path /app/migrations -database ${DATABASE_URL} -verbose up

migrate-down: ## Rollback database migrations
	@docker-compose run --rm api migrate -path /app/migrations -database ${DATABASE_URL} -verbose down
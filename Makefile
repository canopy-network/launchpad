# Launchpad API Makefile

.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt vet tui

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
build: ## Build the application
	go build -o bin/launchpad .

run: ## Run the launchpad locally - requires postgres running
	export $$(grep -v '^#' .env | xargs) && go run .

test: ## Run unit tests only (excludes integration tests)
	go test -v ./...

test-integration: ## Run integration tests only
	go test -v -tags=integration ./tests/integration/...

test-all: ## Run all tests (unit + integration)
	go test -v -tags=integration ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Code quality commands
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	golangci-lint run

# Docker commands
docker-build: ## Build Docker image
	docker build -t launchpad-api .

docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop services with Docker Compose
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

docker-rebuild: ## Rebuild and restart services
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

docker-remove-postgres-all: ## Remove postgres container and volumes
	docker compose stop postgres
	docker compose rm postgres
	docker volume rm -f launchpad_postgres_data

# Database commands
psql: ## Connect to PostgreSQL database via psql
	PGPASSWORD=launchpad123 psql -h localhost -p 5432 -U launchpad -d launchpad

migrate-setup: ## Initialize database with Atlas migrations
	@echo "Setting up database with Atlas migrations..."
	atlas migrate apply --env local
	@echo "Database schema applied successfully!"

load-fixtures: ## Load fixture data into database (run migrate-setup first)
	@echo "Loading fixture data..."
	@echo "Note: Run 'make migrate-setup' first to apply schema migrations"
	PGPASSWORD=launchpad123 psql -h localhost -p 5432 -U launchpad -d launchpad -f fixtures/load_all.sql
	@echo "Fixture data loaded successfully!"

db-reset: ## Reset database with fresh schema and fixture data
	@echo "Resetting database..."
	@echo "yes" | make clear-data
	@echo "Waiting for PostgreSQL to start..."
	@sleep 5
	make migrate-setup
	make load-fixtures
	@echo "Database reset complete!"

clear-data: ## Clear all data from database (WARNING: destructive)
	@echo "WARNING: This will delete all data!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || exit 1
	@echo "Clearing database..."
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Database cleared. You may want to run 'make load-fixtures' to reload test data."

migrate-up: ## Apply pending database migrations using Atlas
	@echo "Applying database migrations..."
	atlas migrate apply --env local
	@echo "Migrations applied successfully!"

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	atlas migrate status --env local

migrate-diff: ## Generate new migration file from schema changes
	@echo "Generating migration diff..."
	@read -p "Enter migration name: " name && atlas migrate diff $$name --env local

migrate-validate: ## Validate migration files
	@echo "Validating migrations..."
	atlas migrate validate --env local

migrate-down: ## Rollback database migrations (Atlas doesn't support automatic rollback)
	@echo "Atlas doesn't support automatic rollbacks."
	@echo "Please create a new migration with the reverse changes."
	@echo "Use 'make migrate-diff' to generate a new migration."

# Development setup
setup: ## Set up development environment
	go mod download
	cp .env.example .env
	@echo "Development environment set up!"
	@echo "Edit .env file with your configuration"

# API testing and chain management
api-test: ## Test API endpoints (basic health and chains)
	@echo "Testing health endpoint..."
	@curl -s http://localhost:3000/health | jq .
	@echo "\nTesting chains endpoint (requires auth header)..."
	@curl -s -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" http://localhost:3000/api/v1/chains | jq .

test-api-full: ## Run comprehensive API test suite
	@./scripts/test_api.sh

create-chain: ## Create a new chain interactively
	@./scripts/create_chain.sh

chain-lifecycle: ## Demonstrate complete chain lifecycle
	@./scripts/chain_lifecycle.sh

chain-quick: ## Quick chain creation with default values
	@./scripts/create_chain.sh "Quick Test Chain" "QUICK" "template-basic-001"

# Production commands
build-prod: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/launchpad .

deploy: build-prod ## Deploy (placeholder)
	@echo "Deployment commands would go here"

tui: ## Run TUI
	@go run cmd/tui/*.go

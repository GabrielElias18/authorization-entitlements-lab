# SpiceDB Model Testing Makefile

.PHONY: help setup deps test test-verbose test-clean docker-up docker-down docker-logs clean install-deps

# Default target
help: ## Show this help message
	@echo "SpiceDB Model Test Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Setup and Dependencies
setup: docker-up deps migrate load-data ## Complete setup: start containers, install deps, migrate DB, load data
	@echo "✅ Setup complete! Ready for testing."

deps: ## Download Go dependencies
	@echo "📦 Installing Go dependencies..."
	@cd tests/go && go mod tidy && go mod download

install-deps: ## Install Go dependencies (alias for deps)
	@$(MAKE) deps

# Docker Management  
docker-up: ## Start SpiceDB and PostgreSQL containers
	@echo "🐳 Starting Docker containers..."
	@cd spicedb-config && docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10

docker-down: ## Stop Docker containers
	@echo "🛑 Stopping Docker containers..."
	@cd spicedb-config && docker-compose down

docker-logs: ## Show Docker container logs
	@echo "📋 Container logs:"
	@cd spicedb-config && docker-compose logs -f

docker-restart: docker-down docker-up ## Restart Docker containers

# Database Operations
migrate: ## Run SpiceDB database migrations
	@echo "🔄 Running database migrations..."
	@docker exec spicedb-config-spicedb-1 spicedb migrate head --datastore-engine postgres --datastore-conn-uri "postgres://spicedb:spicedb@postgres:5432/spicedb?sslmode=disable"

load-schema: ## Load schema into SpiceDB
	@echo "📋 Loading schema..."
	@zed schema write model.zaml

load-data: load-schema ## Load schema and relationship data
	@echo "📊 Loading relationship data..."
	@./load_relationships.sh

reset-data: ## Clear and reload all data
	@echo "🔄 Resetting data..."
	@./reset.sh

# Testing
test: ## Run Go unit tests
	@echo "🧪 Running Go unit tests..."
	@cd tests/go && go test -v ./...

test-verbose: ## Run Go tests with verbose output
	@echo "🧪 Running Go tests (verbose)..."
	@cd tests/go && go test -v -count=1 ./...

test-race: ## Run Go tests with race detection
	@echo "🧪 Running Go tests with race detection..."
	@cd tests/go && go test -race -v ./...

test-bash: ## Run original bash-based tests
	@echo "🧪 Running bash tests..."
	@bash run_permission_checks.sh

test-all: test test-bash ## Run both Go and bash tests

test-clean: reset-data test ## Reset data and run tests

# Development
fmt: ## Format Go code
	@echo "🎨 Formatting Go code..."
	@cd tests/go && go fmt ./...

lint: ## Run Go linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	@cd tests/go && golangci-lint run

build: ## Build the Go application
	@echo "🔨 Building Go application..."
	@cd tests/go && go build -o spicedb-test ./...

# Cleanup
clean: ## Clean up build artifacts and containers
	@echo "🧹 Cleaning up..."
	@rm -f tests/go/spicedb-test
	@docker system prune -f

clean-all: docker-down clean ## Stop containers and clean everything

# CI/Development Workflow
ci: setup test ## Full CI workflow: setup and test
	@echo "✅ CI workflow completed successfully!"

dev: docker-up load-data ## Quick development setup
	@echo "✅ Development environment ready!"

# Database inspection
inspect-schema: ## Show current SpiceDB schema
	@echo "📋 Current schema:"
	@zed schema read

inspect-data: ## Show current relationships
	@echo "📊 Current relationships:"
	@zed relationship read

inspect-all: inspect-schema inspect-data ## Show schema and data

# Performance testing (future)
benchmark: ## Run benchmark tests (placeholder)
	@echo "⚡ Running benchmarks..."
	@cd tests/go && go test -bench=. -benchmem ./...

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@cd tests/go && go doc ./...
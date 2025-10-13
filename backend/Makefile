# ============================================================================
# FILE: backend/Makefile
# UPDATED: Added worker commands
# ============================================================================

.PHONY: help run build test clean migrate dev

# Binary names
BINARY_NAME=socialqueue-api
WORKER_BINARY=socialqueue-worker

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Directories
CMD_DIR=./cmd/api
WORKER_DIR=./cmd/worker
BUILD_DIR=./bin
MIGRATION_DIR=./migrations

# Database
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= socialqueue
DB_PASSWORD ?= socialqueue_dev_password
DB_NAME ?= socialqueue_dev
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Colors
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

##@ General

help: ## Display this help
	@echo "$(BLUE)SocialQueue Backend Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

run: ## Run the API server
	@echo "$(BLUE)Starting API server...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Warning: .env file not found. Creating from .env.example...$(NC)"; \
		cp .env.example .env; \
	fi
	$(GORUN) $(CMD_DIR)

run-worker: ## Run the worker process
	@echo "$(BLUE)Starting worker process...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Error: .env file not found. Please create it from .env.example$(NC)"; \
		exit 1; \
	fi
	$(GORUN) $(WORKER_DIR)

dev: ## Run with hot reload using Air
	@echo "$(BLUE)Starting with hot reload...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Warning: .env file not found. Creating from .env.example...$(NC)"; \
		cp .env.example .env; \
	fi
	air

build: ## Build the API binary
	@echo "$(BLUE)Building API binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(CMD_DIR)
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-worker: ## Build the worker binary
	@echo "$(BLUE)Building worker binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WORKER_BINARY) -v $(WORKER_DIR)
	@echo "$(GREEN)✓ Worker binary built: $(BUILD_DIR)/$(WORKER_BINARY)$(NC)"

build-all: build build-worker ## Build both API and worker binaries
	@echo "$(GREEN)✓ All binaries built successfully!$(NC)"

##@ Testing

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -v -tags=integration ./tests/integration/...
	@echo "$(GREEN)✓ Integration tests completed$(NC)"

##@ Database

migrate: migrate-up ## Alias for migrate-up

migrate-up: ## Run all pending migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Error: .env file not found. Please create it from .env.example$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations completed$(NC)"

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Rollback completed$(NC)"

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Usage: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" force $(VERSION)

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_posts_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=your_migration_name$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)$(NC)"
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created in $(MIGRATION_DIR)$(NC)"

migrate-status: ## Show migration status
	@echo "$(BLUE)Migration status:$(NC)"
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" version

##@ Code Generation

sqlc: ## Generate SQLC code
	@echo "$(BLUE)Generating SQLC code...$(NC)"
	sqlc generate
	@echo "$(GREEN)✓ SQLC code generated$(NC)"

##@ Dependencies

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GOCMD) get -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

deps-verify: ## Verify dependencies
	@echo "$(BLUE)Verifying dependencies...$(NC)"
	$(GOMOD) verify
	@echo "$(GREEN)✓ Dependencies verified$(NC)"

##@ Code Quality

tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Modules tidied$(NC)"

fmt: format ## Alias for format

format: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run
	@echo "$(GREEN)✓ Linting completed$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)✓ Vet completed$(NC)"

check: fmt lint vet test ## Format, lint, vet, and test
	@echo "$(GREEN)✓ All checks passed!$(NC)"

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-cache: ## Clean Go build cache
	@echo "$(BLUE)Cleaning build cache...$(NC)"
	$(GOCMD) clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Cache cleaned$(NC)"

##@ Worker Management (Docker)

worker-logs: ## Tail worker logs (requires Docker)
	@echo "$(BLUE)Tailing worker logs...$(NC)"
	docker-compose logs -f worker

worker-restart: ## Restart worker service (requires Docker)
	@echo "$(YELLOW)Restarting worker...$(NC)"
	docker-compose restart worker
	@echo "$(GREEN)✓ Worker restarted$(NC)"

worker-stop: ## Stop worker service (requires Docker)
	@echo "$(YELLOW)Stopping worker...$(NC)"
	docker-compose stop worker
	@echo "$(GREEN)✓ Worker stopped$(NC)"

worker-start: ## Start worker service (requires Docker)
	@echo "$(BLUE)Starting worker...$(NC)"
	docker-compose start worker
	@echo "$(GREEN)✓ Worker started$(NC)"

##@ Quick Actions

fresh: clean build-all ## Clean and build everything
	@echo "$(GREEN)✓ Fresh build complete!$(NC)"

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	@echo "Installing Air for hot reload..."
	@go install github.com/air-verse/air@latest
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Installing SQLC..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "$(GREEN)✓ Development tools installed!$(NC)"

##@ Info

version: ## Show Go version
	@$(GOCMD) version

env: ## Show environment variables
	@echo "$(BLUE)Environment:$(NC)"
	@echo "  DB_HOST:     $(DB_HOST)"
	@echo "  DB_PORT:     $(DB_PORT)"
	@echo "  DB_NAME:     $(DB_NAME)"
	@echo "  DB_USER:     $(DB_USER)"
	@echo "  BUILD_DIR:   $(BUILD_DIR)"
	@echo "  API_BINARY:  $(BINARY_NAME)"
	@echo "  WORKER_BIN:  $(WORKER_BINARY)"

# ##@ Quick Actions

# fresh-start: docker-down docker-clean docker-up migrate-up run ## Fresh start (clean + migrate + run)

# reset-db: docker-down docker-clean docker-up migrate-up ## Reset database

# quick-test: tidy test ## Tidy and test

# cd /mnt/e/upwork/social-queue/backend

# # Drop database (remove -it flag for non-interactive)
# docker exec socialqueue-postgres psql -U socialqueue -d postgres -c "DROP DATABASE IF EXISTS socialqueue_dev;"

# # Create fresh database
# docker exec socialqueue-postgres psql -U socialqueue -d postgres -c "CREATE DATABASE socialqueue_dev;"

# # Verify it's empty
# docker exec -it socialqueue-postgres psql -U socialqueue -d socialqueue_dev -c "\dt"
# # Should show: "Did not find any relations"
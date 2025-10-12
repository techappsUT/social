# path: backend/Makefile

.PHONY: help run build test clean migrate dev

# Binary names
BINARY_NAME=socialqueue-api
WORKER_BINARY=socialqueue-worker

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Directories
CMD_DIR=./cmd/api
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
	$(GOCMD) run $(CMD_DIR)

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
	$(GOBUILD) -o $(BUILD_DIR)/$(WORKER_BINARY) -v ./cmd/worker
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(WORKER_BINARY)$(NC)"

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

##@ Database

migrate: ## Run all pending migrations
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

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Cleaned$(NC)"

tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Modules tidied$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run
	@echo "$(GREEN)✓ Linting completed$(NC)"

##@ Quick Actions

fresh: clean build ## Clean and build
	@echo "$(GREEN)✓ Fresh build complete!$(NC)"

check: fmt lint test ## Format, lint, and test
	@echo "$(GREEN)✓ All checks passed!$(NC)"

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
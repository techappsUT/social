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
		echo "$(YELLOW)Warning: .env file not found. Using defaults.$(NC)"; \
	fi
	$(GOCMD) run $(CMD_DIR)/main.go

dev: ## Run with air for hot reload
	@echo "$(BLUE)Starting with hot reload...$(NC)"
	@which air > /dev/null || (echo "$(YELLOW)Installing air...$(NC)" && go install github.com/cosmtrek/air@latest)
	air

watch: dev ## Alias for dev

##@ Build

build: ## Build the API binary
	@echo "$(BLUE)Building API binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-linux: ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux -v $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Linux binary built$(NC)"

##@ Testing

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	$(GOTEST) -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

##@ Database

migrate-up: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	@which migrate > /dev/null || (echo "$(YELLOW)Installing migrate...$(NC)" && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations completed$(NC)"

migrate-down: ## Rollback last migration
	@echo "$(BLUE)Rolling back migration...$(NC)"
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Rolled back$(NC)"

migrate-status: ## Check migration status
	@echo "$(BLUE)Migration status:$(NC)"
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" version

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_posts_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(YELLOW)Error: NAME is required. Usage: make migrate-create NAME=your_migration_name$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)$(NC)"
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created in $(MIGRATION_DIR)$(NC)"

##@ Docker

docker-up: ## Start Docker services
	@echo "$(BLUE)Starting Docker services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"

docker-down: ## Stop Docker services
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

docker-logs: ## View Docker logs
	docker-compose logs -f

docker-clean: ## Clean Docker volumes
	@echo "$(BLUE)Cleaning Docker volumes...$(NC)"
	docker-compose down -v
	@echo "$(GREEN)✓ Volumes cleaned$(NC)"

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
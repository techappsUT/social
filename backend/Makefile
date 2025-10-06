# path: backend/Makefile

.PHONY: help run build test test-coverage test-integration lint format clean migrate migrate-create migrate-up migrate-down migrate-status seed dev docker-build

# Binary name
BINARY_NAME=socialqueue-api
WORKER_BINARY=socialqueue-worker

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Directories
CMD_DIR=./cmd/api
WORKER_CMD_DIR=./cmd/worker
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

help: ## Display this help message
	@echo "$(BLUE)Backend (Go) - Development Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

run: ## Run the API server
	@echo "$(BLUE)Starting API server...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Warning: .env file not found. Using defaults.$(NC)"; \
	fi
	$(GOCMD) run $(CMD_DIR)/main.go

run-worker: ## Run the background worker
	@echo "$(BLUE)Starting background worker...$(NC)"
	$(GOCMD) run $(WORKER_CMD_DIR)/main.go

dev: ## Run with air for hot reload (requires air: go install github.com/cosmtrek/air@latest)
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

build-worker: ## Build the worker binary
	@echo "$(BLUE)Building worker binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WORKER_BINARY) -v $(WORKER_CMD_DIR)/main.go
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(WORKER_BINARY)$(NC)"

build-all: build build-worker ## Build all binaries

build-linux: ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux -v $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Linux binary built$(NC)"

##@ Testing

test: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GOTEST) -v -race -timeout 30s ./...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -v -race -tags=integration -timeout 5m ./...

test-short: ## Run short tests only
	@echo "$(BLUE)Running short tests...$(NC)"
	$(GOTEST) -v -race -short ./...

bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

lint: ## Run golangci-lint
	@echo "$(BLUE)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)Installing golangci-lint...$(NC)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout 5m

lint-fix: ## Run golangci-lint with autofix
	@echo "$(BLUE)Running linter with autofix...$(NC)"
	golangci-lint run --fix --timeout 5m

format: ## Format code using gofmt
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) -w -s .
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOCMD) vet ./...

tidy: ## Run go mod tidy
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

verify: ## Verify dependencies
	@echo "$(BLUE)Verifying dependencies...$(NC)"
	$(GOMOD) verify

download: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download

##@ Database Migrations

migrate: migrate-up ## Run all migrations (alias for migrate-up)

migrate-create: ## Create a new migration (usage: make migrate-create name=create_posts_table)
	@if [ -z "$(name)" ]; then \
		echo "$(YELLOW)Usage: make migrate-create name=create_posts_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(name)...$(NC)"
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	touch $(MIGRATION_DIR)/$${timestamp}_$(name).up.sql; \
	touch $(MIGRATION_DIR)/$${timestamp}_$(name).down.sql; \
	echo "$(GREEN)✓ Created:$(NC)"; \
	echo "  • $(MIGRATION_DIR)/$${timestamp}_$(name).up.sql"; \
	echo "  • $(MIGRATION_DIR)/$${timestamp}_$(name).down.sql"

migrate-up: ## Run all up migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	@which migrate > /dev/null || (echo "$(YELLOW)Installing golang-migrate...$(NC)" && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations complete$(NC)"

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1

migrate-down-all: ## Rollback all migrations (WARNING: Deletes all data!)
	@echo "$(YELLOW)⚠️  WARNING: This will rollback ALL migrations!$(NC)"
	@printf "Are you sure? [y/N] "; \
	read REPLY; \
	case "$REPLY" in \
		[Yy]*) \
			migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down -all; \
			echo "$(GREEN)✓ All migrations rolled back$(NC)"; \
			;; \
		*) \
			echo "$(YELLOW)Cancelled.$(NC)"; \
			;; \
	esac

migrate-status: ## Show migration status
	@echo "$(BLUE)Migration status:$(NC)"
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force version=20240101120000)
	@if [ -z "$(version)" ]; then \
		echo "$(YELLOW)Usage: make migrate-force version=20240101120000$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" force $(version)

migrate-goto: ## Go to specific migration (usage: make migrate-goto version=20240101120000)
	@if [ -z "$(version)" ]; then \
		echo "$(YELLOW)Usage: make migrate-goto version=20240101120000$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" goto $(version)

##@ Database

db-shell: ## Open PostgreSQL shell
	@echo "$(BLUE)Opening database shell...$(NC)"
	docker exec -it socialqueue-postgres psql -U socialqueue -d socialqueue_dev

db-create: ## Create database
	@echo "$(BLUE)Creating database...$(NC)"
	docker exec -it socialqueue-postgres createdb -U socialqueue socialqueue_dev
	@echo "$(GREEN)✓ Database created$(NC)"

db-drop: ## Drop database (WARNING: Deletes all data!)
	@echo "$(YELLOW)⚠️  WARNING: This will delete the database!$(NC)"
	@printf "Are you sure? [y/N] "; \
	read REPLY; \
	case "$REPLY" in \
		[Yy]*) \
			docker exec -it socialqueue-postgres dropdb -U socialqueue socialqueue_dev; \
			echo "$(GREEN)✓ Database dropped$(NC)"; \
			;; \
		*) \
			echo "$(YELLOW)Cancelled.$(NC)"; \
			;; \
	esac

db-reset: db-drop db-create migrate ## Reset database (drop, create, migrate)

seed: ## Seed database with sample data
	@echo "$(BLUE)Seeding database...$(NC)"
	$(GOCMD) run ./scripts/seed.go
	@echo "$(GREEN)✓ Database seeded$(NC)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t socialqueue-backend:latest .
	@echo "$(GREEN)✓ Image built$(NC)"

docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 8080:8080 --env-file .env socialqueue-backend:latest

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

clean-cache: ## Clean Go cache
	@echo "$(BLUE)Cleaning Go cache...$(NC)"
	$(GOCMD) clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Cache cleaned$(NC)"

##@ Tools

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Tools installed$(NC)"

swag: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger docs...$(NC)"
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/api/main.go -o ./docs
	@echo "$(GREEN)✓ Swagger docs generated$(NC)"

##@ CI/CD

ci: lint test ## Run CI checks (lint + test)
	@echo "$(GREEN)✓ CI checks passed$(NC)"

pre-commit: format lint test-short ## Run pre-commit checks
	@echo "$(GREEN)✓ Pre-commit checks passed$(NC)"

security-check: ## Run security checks
	@echo "$(BLUE)Running security checks...$(NC)"
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -quiet ./...
	@echo "$(GREEN)✓ Security check complete$(NC)"
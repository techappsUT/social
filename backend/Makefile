# path: backend/Makefile
# 🔄 REFACTORED - Added comprehensive SQLC commands

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ SQLC

.PHONY: sqlc-generate
sqlc-generate: ## Generate SQLC code from SQL queries
	@echo "🔧 Generating SQLC code..."
	@which sqlc > /dev/null || (echo "Installing sqlc..." && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
	sqlc generate
	@echo "✅ SQLC code generated successfully"

.PHONY: sqlc-verify
sqlc-verify: ## Verify SQLC configuration
	@echo "🔍 Verifying SQLC configuration..."
	sqlc verify
	@echo "✅ SQLC configuration valid"

.PHONY: sqlc-vet
sqlc-vet: ## Run SQLC vet checks
	@echo "🔍 Running SQLC vet..."
	sqlc vet
	@echo "✅ SQLC vet passed"

.PHONY: sqlc-compile
sqlc-compile: ## Compile SQLC queries without generating code
	@echo "🔍 Compiling SQLC queries..."
	sqlc compile
	@echo "✅ SQLC queries compiled successfully"

##@ Database

.PHONY: db-up
db-up: ## Start database services
	@echo "🐳 Starting database services..."
	docker-compose up -d postgres redis
	@sleep 3
	@echo "✅ Database services started"

.PHONY: db-down
db-down: ## Stop database services
	@echo "🛑 Stopping database services..."
	docker-compose down
	@echo "✅ Database services stopped"

.PHONY: db-logs
db-logs: ## Show database logs
	docker-compose logs -f postgres

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell
	docker exec -it socialqueue-postgres psql -U socialqueue -d socialqueue_dev

##@ Migrations

.PHONY: migrate-up
migrate-up: db-up ## Run all pending migrations
	@echo "⬆️  Running migrations..."
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path migrations -database "postgresql://socialqueue:socialqueue_dev_password@localhost:5432/socialqueue_dev?sslmode=disable" up
	@echo "✅ Migrations completed"

.PHONY: migrate-down
migrate-down: ## Rollback one migration
	@echo "⬇️  Rolling back migration..."
	migrate -path migrations -database "postgresql://socialqueue:socialqueue_dev_password@localhost:5432/socialqueue_dev?sslmode=disable" down 1
	@echo "✅ Migration rolled back"

.PHONY: migrate-status
migrate-status: ## Show migration status
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path migrations -database "postgresql://socialqueue:socialqueue_dev_password@localhost:5432/socialqueue_dev?sslmode=disable" version

##@ Setup

.PHONY: setup
setup: db-up migrate-up sqlc-generate ## Complete setup (DB + Migrations + SQLC)
	@echo "✅ Setup complete! You can now run 'make dev'"

.PHONY: setup-clean
setup-clean: db-down ## Clean setup and start fresh
	@echo "🧹 Cleaning up..."
	docker volume rm socialqueue_postgres-data 2>/dev/null || true
	docker volume rm socialqueue_redis-data 2>/dev/null || true
	@echo "✅ Cleanup complete"
	@$(MAKE) setup

##@ Development

.PHONY: dev
dev: ## Run development server
	@echo "🚀 Starting development server..."
	go run cmd/server/main.go

.PHONY: dev-api
dev-api: ## Run API server (legacy)
	@echo "🚀 Starting API server..."
	go run cmd/api/main.go

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	goimports -w .

##@ Build

.PHONY: build
build: ## Build binary
	@echo "🔨 Building binary..."
	go build -o bin/socialqueue cmd/server/main.go
	@echo "✅ Binary built: bin/socialqueue"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
	@echo "✅ Cleanup complete"
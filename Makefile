# path: Makefile

.PHONY: help setup install build up down restart logs clean test migrate seed dev prod backend frontend db-shell db-reset

# Colors for pretty output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

##@ General

help: ## Display this help message
	@echo "$(BLUE)SocialQueue - Development Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup & Installation

setup: ## Initial project setup (install deps, setup env files)
	@echo "$(BLUE)Setting up SocialQueue...$(NC)"
	@make install
	@make setup-env
	@echo "$(GREEN)✓ Setup complete!$(NC)"

install: ## Install all dependencies (backend + frontend)
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@make backend-install
	@make frontend-install
	@echo "$(GREEN)✓ Dependencies installed!$(NC)"

setup-env: ## Copy environment files if they don't exist
	@if [ ! -f backend/.env ]; then \
		echo "$(YELLOW)Creating backend/.env from .env.example...$(NC)"; \
		cp backend/.env.example backend/.env; \
	fi
	@if [ ! -f frontend/.env.local ]; then \
		echo "$(YELLOW)Creating frontend/.env.local from .env.example...$(NC)"; \
		cp frontend/.env.example frontend/.env.local; \
	fi
	@echo "$(GREEN)✓ Environment files ready!$(NC)"

##@ Docker Services

up: ## Start all services (detached)
	@echo "$(BLUE)Starting all services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started!$(NC)"
	@make status

up-build: ## Build and start all services
	@echo "$(BLUE)Building and starting all services...$(NC)"
	docker-compose up -d --build
	@echo "$(GREEN)✓ Services built and started!$(NC)"

down: ## Stop all services
	@echo "$(YELLOW)Stopping all services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped!$(NC)"

restart: ## Restart all services
	@echo "$(YELLOW)Restarting services...$(NC)"
	@make down
	@make up

stop: ## Stop all services (alias for down)
	@make down

ps: status ## Show running services (alias for status)

status: ## Show status of all services
	@echo "$(BLUE)Service Status:$(NC)"
	@docker-compose ps

logs: ## Tail logs for all services
	docker-compose logs -f

logs-backend: ## Tail backend logs
	docker-compose logs -f backend

logs-frontend: ## Tail frontend logs
	docker-compose logs -f frontend

logs-postgres: ## Tail postgres logs
	docker-compose logs -f postgres

##@ Database Commands

db-up: ## Start only database services (postgres + redis)
	@echo "$(BLUE)Starting database services...$(NC)"
	docker-compose up -d postgres redis adminer redis-commander
	@echo "$(GREEN)✓ Database services started!$(NC)"

db-down: ## Stop database services
	@echo "$(YELLOW)Stopping database services...$(NC)"
	docker-compose stop postgres redis adminer redis-commander
	@echo "$(GREEN)✓ Database services stopped!$(NC)"

db-shell: ## Open PostgreSQL shell
	@echo "$(BLUE)Opening database shell...$(NC)"
	docker-compose exec postgres psql -U socialqueue -d socialqueue_dev

db-reset: ## Reset database (WARNING: Deletes all data!)
	@echo "$(RED)⚠️  WARNING: This will delete all data!$(NC)"
	@printf "Are you sure? [y/N] "; \
	read REPLY; \
	case "$REPLY" in \
		[Yy]*) \
			echo "$(YELLOW)Resetting database...$(NC)"; \
			docker-compose down -v; \
			docker-compose up -d postgres redis; \
			echo "$(YELLOW)Waiting for PostgreSQL to be ready...$(NC)"; \
			sleep 5; \
			make migrate; \
			echo "$(GREEN)✓ Database reset complete!$(NC)"; \
			;; \
		*) \
			echo "$(YELLOW)Cancelled.$(NC)"; \
			;; \
	esac

migrate: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	cd backend && make migrate
	@echo "$(GREEN)✓ Migrations complete!$(NC)"

migrate-create: ## Create a new migration (usage: make migrate-create name=create_posts_table)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: name is required$(NC)"; \
		echo "Usage: make migrate-create name=create_posts_table"; \
		exit 1; \
	fi
	cd backend && make migrate-create name=$(name)

seed: ## Seed database with sample data
	@echo "$(BLUE)Seeding database...$(NC)"
	cd backend && make seed
	@echo "$(GREEN)✓ Database seeded!$(NC)"

##@ Development

dev: ## Start development environment (databases + live reload)
	@echo "$(BLUE)Starting development environment...$(NC)"
	@make db-up
	@echo "$(YELLOW)Waiting for databases to be ready...$(NC)"
	@sleep 5
	@make migrate
	@echo ""
	@echo "$(GREEN)✓ Development environment ready!$(NC)"
	@echo ""
	@echo "$(BLUE)Database Admin Tools:$(NC)"
	@echo "  • Adminer: http://localhost:8081"
	@echo "  • Redis Commander: http://localhost:8082"
	@echo ""
	@echo "$(YELLOW)To start backend:$(NC)  cd backend && make run"
	@echo "$(YELLOW)To start frontend:$(NC) cd frontend && npm run dev"
	@echo ""

dev-full: ## Start full development environment (all services in Docker)
	@make up-build
	@echo ""
	@echo "$(GREEN)✓ Full development environment started!$(NC)"
	@echo ""
	@echo "$(BLUE)Services:$(NC)"
	@echo "  • Frontend: http://localhost:3000"
	@echo "  • Backend API: http://localhost:8080"
	@echo "  • Adminer: http://localhost:8081"
	@echo "  • Redis Commander: http://localhost:8082"
	@echo ""

##@ Backend Commands

backend-install: ## Install backend dependencies
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	cd backend && go mod download
	@echo "$(GREEN)✓ Backend dependencies installed!$(NC)"

backend-run: ## Run backend server locally
	cd backend && make run

backend-test: ## Run backend tests
	cd backend && make test

backend-test-coverage: ## Run backend tests with coverage
	cd backend && make test-coverage

backend-lint: ## Lint backend code
	cd backend && make lint

backend-build: ## Build backend binary
	cd backend && make build

##@ Frontend Commands

frontend-install: ## Install frontend dependencies
	@echo "$(BLUE)Installing frontend dependencies...$(NC)"
	cd frontend && pnpm install
	@echo "$(GREEN)✓ Frontend dependencies installed!$(NC)"

frontend-run: ## Run frontend dev server locally
	cd frontend && pnpm dev

frontend-build: ## Build frontend for production
	cd frontend && pnpm build

frontend-test: ## Run frontend tests
	cd frontend && pnpm test

frontend-lint: ## Lint frontend code
	cd frontend && pnpm lint

frontend-format: ## Format frontend code
	cd frontend && pnpm format

##@ Testing

test: ## Run all tests (backend + frontend)
	@echo "$(BLUE)Running all tests...$(NC)"
	@make backend-test
	@make frontend-test
	@echo "$(GREEN)✓ All tests passed!$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	cd backend && make test-integration
	@echo "$(GREEN)✓ Integration tests passed!$(NC)"

##@ Build & Deploy

build: ## Build all services for production
	@echo "$(BLUE)Building all services...$(NC)"
	@make backend-build
	@make frontend-build
	@echo "$(GREEN)✓ Build complete!$(NC)"

prod: ## Start production environment
	@echo "$(BLUE)Starting production environment...$(NC)"
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "$(GREEN)✓ Production environment started!$(NC)"

##@ Cleanup

clean: ## Clean all generated files and caches
	@echo "$(YELLOW)Cleaning generated files...$(NC)"
	@make backend-clean
	@make frontend-clean
	@echo "$(GREEN)✓ Cleanup complete!$(NC)"

backend-clean: ## Clean backend build artifacts
	cd backend && make clean

frontend-clean: ## Clean frontend build artifacts
	cd frontend && rm -rf .next node_modules/.cache .pnpm-store

clean-volumes: ## Remove all Docker volumes (WARNING: Deletes data!)
	@echo "$(RED)⚠️  WARNING: This will delete all Docker volumes!$(NC)"
	@printf "Are you sure? [y/N] "; \
	read REPLY; \
	case "$REPLY" in \
		[Yy]*) \
			docker-compose down -v; \
			echo "$(GREEN)✓ Volumes removed!$(NC)"; \
			;; \
		*) \
			echo "$(YELLOW)Cancelled.$(NC)"; \
			;; \
	esac

clean-all: ## Clean everything (files, volumes, images)
	@echo "$(RED)⚠️  WARNING: This will remove all data, images, and containers!$(NC)"
	@printf "Are you sure? [y/N] "; \
	read REPLY; \
	case "$REPLY" in \
		[Yy]*) \
			docker-compose down -v --rmi all; \
			make clean; \
			echo "$(GREEN)✓ Everything cleaned!$(NC)"; \
			;; \
		*) \
			echo "$(YELLOW)Cancelled.$(NC)"; \
			;; \
	esac

##@ Utilities

format: ## Format all code (backend + frontend)
	@echo "$(BLUE)Formatting code...$(NC)"
	cd backend && make format
	cd frontend && pnpm format
	@echo "$(GREEN)✓ Code formatted!$(NC)"

lint-all: ## Lint all code (backend + frontend)
	@echo "$(BLUE)Linting all code...$(NC)"
	@make backend-lint
	@make frontend-lint
	@echo "$(GREEN)✓ Linting complete!$(NC)"

health-check: ## Check health of all services
	@echo "$(BLUE)Checking service health...$(NC)"
	@echo -n "Backend API: "
	@curl -sf http://localhost:8080/health > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Down$(NC)"
	@echo -n "Frontend: "
	@curl -sf http://localhost:3000 > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Down$(NC)"
	@echo -n "PostgreSQL: "
	@docker-compose exec -T postgres pg_isready -U socialqueue > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Down$(NC)"
	@echo -n "Redis: "
	@docker-compose exec -T redis redis-cli ping > /dev/null && echo "$(GREEN)✓ Healthy$(NC)" || echo "$(RED)✗ Down$(NC)"

update-deps: ## Update all dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	cd backend && go get -u ./... && go mod tidy
	cd frontend && pnpm update
	@echo "$(GREEN)✓ Dependencies updated!$(NC)"

##@ Documentation

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	@echo "$(YELLOW)Documentation generation not yet implemented$(NC)"

api-docs: ## Open API documentation
	@echo "Opening API docs..."
	@open http://localhost:8080/swagger/index.html || xdg-open http://localhost:8080/swagger/index.html
.PHONY: help dev up down backend frontend migrate db-shell redis-shell minio-shell test build clean

help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: up backend ## Start all dev services

up: ## Start Docker services
	docker compose up -d

down: ## Stop Docker services
	docker compose down

backend: ## Run backend server
	cd backend && go run cmd/server/main.go

frontend: ## Run frontend dev server
	cd frontend && pnpm dev

migrate: ## Run database migrations
	cd backend && go run cmd/migrate/main.go

db-shell: ## Connect to PostgreSQL
	docker compose exec postgres psql -U tiaozao -d tiaozao

redis-shell: ## Connect to Redis CLI
	docker compose exec redis redis-cli

minio-shell: ## Use mc to connect
	docker compose exec minio mc ls

test: ## Run all tests
	cd backend && go test ./...

build: ## Build backend and frontend
	cd backend && go build -o bin/server cmd/server/main.go
	cd frontend && pnpm build

clean: ## Clean build artifacts
	rm -rf backend/bin frontend/dist
	docker compose down -v

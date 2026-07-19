.PHONY: dev build test lint migrate clean help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start local dev environment with Docker Compose
	docker compose up --build

dev-d: ## Start local dev environment in detached mode
	docker compose up --build -d

stop: ## Stop Docker Compose services
	docker compose down

logs: ## View Docker Compose logs
	docker compose logs -f

build: build-backend build-frontend ## Build all

build-backend: ## Build backend
	cd backend && go build -o tmp/kareelio-server ./cmd/server

build-frontend: ## Build frontend
	cd frontend && npm run build

test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	cd backend && go test ./... -v -race -coverprofile=coverage.out

test-frontend: ## Run frontend tests
	cd frontend && npm run test

lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Lint backend
	cd backend && go vet ./...

lint-frontend: ## Lint frontend
	cd frontend && npm run lint

migrate: ## Run database migrations
	cd backend && go run ./cmd/server migrate

clean: ## Clean build artifacts
	rm -rf backend/tmp/ frontend/dist/ frontend/node_modules/
	docker compose down -v

install: install-backend install-frontend ## Install all dependencies

install-backend: ## Install Go dependencies
	cd backend && go mod download

install-frontend: ## Install frontend dependencies
	cd frontend && npm install

release: ## Run semantic-release
	npx semantic-release

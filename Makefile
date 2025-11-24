.PHONY: help install-tools test test-e2e test-coverage watch-test \
        build clean start-server dev docker-up wait-for-db \
        migrate migrate-down lint fmt vet pre-push start-seeder

# Variables
GO := go
BINARY_NAME := bin/server
MIGRATIONS_PATH := internal/database/migrations
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Database variables (can be overridden by env vars)
DB_HOST ?= 127.0.0.1
DB_PORT ?= 3306
DB_USER ?= user
DB_PASSWORD ?= password
DB_NAME ?= dbname
DB_DSN := mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

## Help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-20s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Install Tools: Install necessary Go tools
install-tools:
	@echo "Ensuring Go modules are tidy..."
	@$(GO) mod tidy
	@echo "Installing tools..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2; }
	@command -v migrate >/dev/null 2>&1 || { echo "Installing migrate..."; $(GO) install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; }
	@command -v air >/dev/null 2>&1 || { echo "Installing Air..."; $(GO) install github.com/cosmtrek/air@latest; }
	@command -v gotestsum >/dev/null 2>&1 || { echo "Installing gotestsum..."; $(GO) install gotest.tools/gotestsum@latest; }
	@echo "✅ Tools installed."

## Build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build -o $(BINARY_NAME) ./cmd/server/main.go
	@echo "✅ Build complete."

## Clean: Remove generated files
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) coverage-summary.txt
	@echo "✅ Clean complete."

## Test: Run unit tests
test: install-tools
	@echo "Running tests..."
	@gotestsum --format=short-verbose -- $(shell $(GO) list ./... | grep -v -E '/(cmd|docs|tests)')

## Test E2E: Run end-to-end tests
test-e2e: install-tools
	@echo "Running E2E tests..."
	@gotestsum --format=short-verbose -- ./tests/e2e/...

## Test Coverage: Run tests with coverage
test-coverage: install-tools
	@echo "Running tests with coverage..."
	@gotestsum -- -coverprofile=$(COVERAGE_FILE) $(shell $(GO) list ./... | grep -v -E '/(cmd|docs|tests)')
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tee coverage-summary.txt
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✅ Coverage report generated."

## Watch Test: Watch for changes and run tests
watch-test: install-tools
	@echo "Watching for changes..."
	@reflex -r '\.go$$' -s -- sh -c 'clear && gotestsum --format=short-verbose -- $(shell $(GO) list ./... | grep -v -E "/(cmd|docs|tests)")'

## Docker Up: Start Docker containers
docker-up:
	@echo "📦 Starting Docker containers..."
	@if docker info > /dev/null 2>&1; then \
		docker-compose up -d || { echo '❌ Failed to start containers'; exit 1; }; \
	else \
		echo "❌ Docker is not running."; exit 1; \
	fi

## Wait for DB: Wait for MySQL to be ready
wait-for-db:
	@echo "⏱️  Waiting for MySQL..."
	@retries=0; max_retries=30; \
	until docker exec $$(docker ps -qf "name=mysql") mysqladmin ping -h"127.0.0.1" --silent > /dev/null 2>&1; do \
		printf "🐢"; sleep 1; retries=$$((retries+1)); \
		if [ $$retries -ge $$max_retries ]; then echo "\n❌ MySQL timeout."; exit 1; fi; \
	done
	@echo "\n✅ MySQL is ready."

## Dev: Start server with Air (requires DB)
dev: install-tools
	@echo "⚙️  Starting Air..."
	@air || { echo '❌ Failed to start Air'; exit 1; }

## Start Server: Full dev environment setup
start-server: docker-up wait-for-db dev

## Start Seeder: Seed the database
start-seeder:
	@echo "Seeding database..."
	@$(GO) run ./cmd/seeder/seeder.go
	@echo "✅ Seeding complete."

## Migrate: Run database migrations up
migrate: install-tools
	@echo "🔄 Running migrations up..."
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	migrate -path $(MIGRATIONS_PATH) -database "mysql://$${DB_USERNAME}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_DATABASE}" up
	@echo "✅ Migrations up complete."

## Migrate Down: Revert database migrations
migrate-down: install-tools
	@echo "🔄 Running migrations down..."
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	migrate -path $(MIGRATIONS_PATH) -database "mysql://$${DB_USERNAME}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_DATABASE}" down
	@echo "✅ Migrations down complete."

## Lint: Run linter
lint: install-tools
	@echo "🔎 Running linter..."
	@golangci-lint run --timeout=7m --skip-dirs=cmd,docs,tests ./...

## Fmt: Format code
fmt: install-tools
	@echo "📝 Formatting code..."
	@$(GO) fmt ./...

## Vet: Run go vet
vet: install-tools
	@echo "🔍 Running go vet..."
	@$(GO) vet ./...

## Pre-push: Run all checks before push
pre-push: fmt vet lint test
	@echo "✅ Pre-push checks passed."

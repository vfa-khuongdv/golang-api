.PHONY: help install-tools test test-e2e test-coverage watch-test \
        build clean dev lint fmt vet pre-push

# Variables
GO := go
BINARY_NAME := bin/server
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Core modules for coverage reporting (excluding cmd, models, routes, seeders, mocks)
# These modules represent the business logic that should be tested
CORE_MODULES := ./internal/shared/... ./internal/handlers ./internal/middlewares \
                ./internal/repositories ./internal/services ./pkg/...

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
	@gotestsum -- -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(CORE_MODULES)
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tee coverage-summary.txt
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✅ Coverage report generated."

## Watch Test: Watch for changes and run tests
watch-test: install-tools
	@echo "Watching for changes..."
	@reflex -r '\.go$$' -s -- sh -c 'clear && gotestsum --format=short-verbose -- $(shell $(GO) list ./... | grep -v -E "/(cmd|docs|tests)")'

## Dev: Start server with Air (requires DB)
dev: install-tools
	@echo "⚙️  Starting Air..."
	@air || { echo '❌ Failed to start Air'; exit 1; }

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

.PHONY: install-tools test-coverage test watch-test start-server start-seeder migrate

install-tools:
	@echo "Ensuring Go modules are tidy..."
	@go mod tidy

	@echo "Installing Go toolchain dependencies..."

	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Installing migrate CLI..."; \
		go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		echo "✅ migrate CLI installed"; \
	else \
		echo "✅ migrate already installed"; \
	fi

	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing Air (live reload)..."; \
		go install github.com/cosmtrek/air@latest; \
		echo "✅ Air installed"; \
	else \
		echo "✅ air already installed"; \
	fi

	@if ! command -v gotestsum >/dev/null 2>&1; then \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
		echo "✅ gotestsum installed"; \
	else \
		echo "✅ gotestsum already installed"; \
	fi

	@echo "Installing all Go dependencies (go install ./...)"
	@go install ./...

	@echo "✅ All tools and packages installed."

test: install-tools
	@echo "Running tests with gotestsum..."
	@gotestsum --format=short-verbose -- $(shell go list ./... | grep -v -E '/(cmd|docs|tests)')
	@echo "✅ Tests completed."

test-coverage: install-tools
	@echo "Running tests with coverage..."
	@gotestsum -- -coverprofile=coverage.out $(shell go list ./... | grep -v -E '/(cmd|docs|tests)')
	@echo "Generating coverage report..."
	@go tool cover -func=coverage.out | tee coverage-summary.txt
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage summary written to coverage-summary.txt"
	@echo "✅ Coverage HTML report generated at coverage.html"

watch-test: install-tools
	@echo "Watching for changes and running tests..."
	@reflex -r '\.go$$' -s -- sh -c 'clear && gotestsum --format=short-verbose -- $(shell go list ./... | grep -v -E "/(cmd|docs|tests)")'

start-server: install-tools
	@echo "🚀 Starting development server setup..."

	@echo "🐳 Checking Docker status..."
	@if docker info > /dev/null 2>&1; then \
		echo "✅ Docker is running."; \
	else \
		echo "❌ Docker is not running. Please start Docker Desktop or Docker Engine."; \
		exit 1; \
	fi

	@echo "📦 Starting Docker containers in detached mode..."
	@docker-compose up -d || { echo '❌ Failed to start containers'; exit 1; }

	@echo "⏱️  Waiting for MySQL to be ready..."
	@retries=0; max_retries=30; \
	until docker exec $$(docker ps -qf "name=mysql") \
		mysqladmin ping -h"127.0.0.1" --silent > /dev/null 2>&1; do \
		printf "🐢"; \
		sleep 1; \
		retries=$$((retries+1)); \
		if [ $$retries -ge $$max_retries ]; then \
			echo "\n❌ MySQL did not become ready in time."; \
			exit 1; \
		fi; \
	done
	@echo "\n✅ MySQL is ready."

	@echo "⚙️  Starting server with live reload using Air..."
	@air || { echo '❌ Failed to start the server with Air'; exit 1; }

	@echo "🎉 Server is running and ready for development!"


start-seeder:
	@echo "Seeding the database..."
	go run ./cmd/seeder/seeder.go
	@echo "Database seeding completed"

migrate: install-tools
	@echo "🔄 Running database migrations..."
	@if [ -f .env ]; then \
		echo "📄 Loading environment variables from .env file..."; \
		export $$(grep -v '^#' .env | xargs) && \
		migrate -path internal/database/migrations -database "mysql://$${DB_USERNAME}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_DATABASE}" up; \
	else \
		echo "⚠️  No .env file found, using system environment variables..."; \
		migrate -path internal/database/migrations -database "mysql://$${DB_USERNAME}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_DATABASE}" up; \
	fi
	@echo "✅ Database migrations completed"

# Commands Reference

## Development

```bash
make dev                # Hot reload with Air (needs MySQL and .env)
go run cmd/server/main.go   # Run from the repository root
```

## Build

```bash
make build              # Output: ./bin/server
go build -o bin/server ./cmd/server/main.go
```

## Testing

```bash
make test               # Unit tests with gotestsum (excludes cmd, docs, tests)
make test-coverage      # With coverage report
make test-e2e           # E2E tests
make watch-test         # Re-run tests on change (needs reflex)

# Direct
go test ./... -v --cover
go test ./... -race
go test ./tests/e2e/... -v
```

## Code Quality

```bash
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
make pre-push           # Full check (fmt vet lint test)
make install-tools      # Install golangci-lint, air, gotestsum
make help               # List all targets
```

## Database

```bash
docker-compose up -d mysql mailpit   # Start MySQL and Mailpit (web UI on :8026)
RUN_MIGRATE=true go run cmd/server/main.go   # Apply migrations (internal/database/migrations)
go run cmd/seeder/seeder.go  # Seed sample users (tables must exist)
make encrypt-setting         # Encrypt a secret setting value
```

Create a migration (needs the [migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)):

```bash
migrate create -ext sql -dir internal/database/migrations -seq your_migration_name
```

## Prerequisites

- Go 1.27+
- MySQL 8.0+
- Docker (for local MySQL, phpMyAdmin, and Mailpit)

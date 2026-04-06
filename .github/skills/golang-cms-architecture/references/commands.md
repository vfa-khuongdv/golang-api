# Commands Reference

## Development

```bash
make dev                # Hot reload with Air
go run cmd/server/main.go
```

## Build

```bash
make build              # Output: ./bin/server
go build -o bin/server ./cmd/server/main.go
```

## Testing

```bash
make test               # Unit tests with gotestsum
make test-coverage      # With coverage report
make test-e2e           # E2E tests

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
```

## Database

```bash
docker-compose up -d mysql   # Start MySQL
go run cmd/seeder/seeder.go  # Seed data
```

## Prerequisites

- Go 1.25+
- MySQL 8.0+

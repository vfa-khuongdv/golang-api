# Golang CMS - Copilot Instructions

This is a Go 1.27+ CMS REST API with JWT auth (access + refresh tokens), password reset by email, and clean architecture.
See [.github/skills/golang-cms-architecture/SKILL.md](skills/golang-cms-architecture/SKILL.md) for comprehensive guidelines.

**Key Technologies:** Go 1.27+, Gin, GORM, MySQL, JWT, Testify, Docker

## Build & Validation Commands

**Prerequisites:** Go 1.27+, MySQL 8.0+, Make, Docker (for local MySQL/Mailpit)

**Bootstrap & Setup:**
```bash
# Install dependencies
go mod download

# Configure the environment (DB_*, JWT_KEY and SETTINGS_ENCRYPTION_KEY are required)
cp .env.example .env

# Setup database (requires running MySQL container)
docker-compose up -d mysql

# Migrations are applied when the server starts with RUN_MIGRATE=true
# (SQL files in internal/database/migrations); the seeder only inserts sample users
```

**Build:**
```bash
# Build server binary
make build

# Output: ./bin/server executable
```

**Tests & Validation:**
```bash
# Run all tests with coverage
go test ./... -v --cover

# All tests must pass (unit tests plus tests/e2e). Current coverage:
# - configs:        100.0%
# - handlers:       100.0%
# - middlewares:    100.0%
# - repositories:   98.2%
# - services:       100.0%
# - shared/utils:   99.6%
# - pkg/apperror:   100.0%
# - pkg/logger:     100.0%
# - pkg/mailer:     100.0%
# - pkg/migrator:   96.4%
# CI fails when total coverage of the unit-tested packages drops below 70%.

# Run specific package tests
go test ./internal/handlers -v

# Run specific test
go test ./internal/handlers -run TestUpdateProfile -v

# Run end-to-end tests only
make test-e2e

# Run with race condition detection
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Linting:**
```bash
# Go format check
go fmt ./...

# Go vet analysis
go vet ./...

# Run all linting (requires golangci-lint)
golangci-lint run ./...
```

**Running the Application:**
```bash
# Start database container
docker-compose up -d mysql

# Start server (default port 3000; run from the repository root)
RUN_MIGRATE=true go run cmd/server/main.go

# Seed sample users (after the tables exist)
go run cmd/seeder/seeder.go

```

## Quick Reference

### Layers
- **Handlers** → Parse requests, call services, return responses
- **Services** → Business logic, validation, orchestrate repositories
- **Repositories** → DB operations, implement interfaces
- **Models** → Domain objects with GORM/JSON tags (snake_case)
- **DTOs** → Request/response types with `binding` validation tags (`internal/shared/dto`)

### Key Rules
- Depend on interfaces, not concrete types (`Reader`, `Writer`, `UserRepository`)
- Use apperror for all errors (validation/404/401/409/500); respond with `utils.RespondWithError` / `utils.RespondWithOK`
- Every service and repository method takes `context.Context` first
- JSON tags in snake_case: `created_at`, `user_id`
- Test files: `*_test.go`, grouped with `t.Run()`
- Functions: Get/Create/Update/Delete/Is/Has prefix pattern

## Testing Standards (see TESTING.md for details)

**Framework:** Testify with AAA pattern (Arrange-Act-Assert)
- `require`: Hard assertions (stop on failure)
- `assert`: Soft assertions (continue on failure)
- `mock`: Object mocking for dependencies

**Test Structure:**
```go
func TestUserService(t *testing.T) {
    t.Run("GetProfile - Success", func(t *testing.T) {
        // Arrange: setup and expectations
        mockRepo := new(mocks.MockUserRepository)
        service := services.NewUserService(mockRepo, new(mocks.MockMailerService))
        mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{ID: 1}, nil).Once()

        // Act: execute function
        result, err := service.GetProfile(context.Background(), 1)

        // Assert: verify results
        require.NoError(t, err)
        assert.NotNil(t, result)
        mockRepo.AssertExpectations(t)
    })
}
```

**Coverage Targets:**
- Handlers: 95% | Services: 85% | Repositories: 90% | Middlewares: 85% | Utils: 80%
- Run: `go test ./... --cover`

**Key Rules:**
- Mock external dependencies (repositories, services) with the mocks in `tests/mocks`
- In-memory SQLite for repository unit tests and `tests/e2e`
- Unit tests sit next to the code (`*_test.go`); `tests/` only holds `e2e` and `mocks`
- Both success and failure test cases
- Prefer testify assertions over `t.Errorf`

## Important Implementation Rules

✓ **Always DO:**
- Use dependency injection
- Handle all errors explicitly
- Use apperror for all errors
- Write tests immediately after code
- Follow AAA pattern in tests
- Mock external dependencies only
- Use meaningful names
- Keep interfaces small (1-3 methods)

✗ **Never DO:**
- Ignore errors silently
- Use global variables
- Mix concerns between layers
- Hardcode config values
- Store passwords in plain text
- Log sensitive information
- Test multiple things per test
- Mix unit and integration tests

## Continuous Integration Checks

When making changes, verify:
1. **All tests pass:** `go test ./... -v --cover`
2. **No formatting issues:** `go fmt ./...`
3. **No vet warnings:** `go vet ./...`
4. **No lint errors:** `golangci-lint run ./...` (if installed)
5. **Server builds:** `go build -o ./bin/server ./cmd/server/main.go`
6. **Migrations apply:** start the server once with `RUN_MIGRATE=true`

## Configuration & Environment

Environment variables (from `.env` file, see `internal/configs/config.go`):
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE` - Database connection
- `JWT_KEY` - Secret key for JWT token signing (min 32 characters)
- `SETTINGS_ENCRYPTION_KEY` - Key encrypting secret settings such as `mail_password` (min 32 characters)
- `PORT` - Server port (default: 3000)
- `RUN_MIGRATE` - Apply SQL migrations on startup when `true` (default: false)
- `STAGE`, `GIN_MODE`, `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `APP_SERVICE`, `APP_VERSION` - Optional, see README

For local development, see `.env.example` in root directory.

Mail (`mail_*`) and frontend (`frontend_url`) settings live in the `settings` table (key/value), not in environment variables. `mail_password` is stored encrypted (`make encrypt-setting`).

## When Adding New Features

1. **Start with the model** - Define domain model with validation tags
2. **Add repository layer** - Define interface, implement GORM ops, write unit tests
3. **Add service layer** - Business logic, validate inputs, orchestrate repos
4. **Add handler layer** - Accept requests, call service, return responses
5. **Add routes** - Register handler in internal/routes/routes.go
6. **Add a migration** - New tables/columns go in `internal/database/migrations` (`*.up.sql` and `*.down.sql`)
7. **Run all tests** - `go test ./... -v --cover` must pass with coverage targets
8. **Update docs** - Update `docs/swagger.json` and the README endpoint list for API changes; update DEVELOPMENT.md and TESTING.md for new patterns

## Debugging Tips

- Use `go test -run TestName -v` for specific tests
- Use `go test -race` for race conditions
- Check `internal/repositories/*_test.go` and `tests/e2e/setup_test.go` for test database setup
- Check `tests/mocks/` for mock examples
- View swagger at `/swagger` (or `/api-docs`) when the server is running and `STAGE` is not `prod`

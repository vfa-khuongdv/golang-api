# Golang CMS - Copilot Instructions

This is a Go 1.21+ CMS REST API with JWT auth, MFA (TOTP), and clean architecture. 
See [.github/skills/golang-cms-architecture/SKILL.md](skills/golang-cms-architecture/SKILL.md) for comprehensive guidelines.

**Key Technologies:** Go 1.21+, Gin, GORM, MySQL, JWT, Testify, Docker

## Build & Validation Commands

**Prerequisites:** Go 1.21+, MySQL 8.0+, Make

**Bootstrap & Setup:**
```bash
# Install dependencies
go mod download

# Setup database (requires running MySQL container)
docker-compose up -d mysql

# Run migrations
go run cmd/seeder/seeder.go
```

**Build:**
```bash
# Build server binary
make build

# Output: ./tmp/server executable
```

**Tests & Validation:**
```bash
# Run all tests with coverage
go test ./... -v --cover

# All tests must pass. Expected: 11 packages passing with 80%-100% coverage
# - handlers:       99.0%
# - repositories:   95.8%
# - services:       83.8%
# - utils:          84.6%
# - pkg/apperror:   100.0%
# - pkg/logger:     83.3%
# - pkg/mailer:     100.0%
# - pkg/migrator:   96.0%

# Run specific package tests
go test ./internal/handlers -v

# Run specific test
go test ./internal/handlers -run TestUserHandler -v

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

# Run migrations and seeding
go run cmd/seeder/seeder.go

# Start server (default port 3000)
go run cmd/server/main.go

# Server runs on http://localhost:3000
```

## Quick Reference

### Layers
- **Handlers** → Parse requests, call services, return responses
- **Services** → Business logic, validation, orchestrate repositories
- **Repositories** → DB operations, implement interfaces
- **Models** → Domain objects with GORM/JSON tags (snake_case)

### Key Rules
- Depend on interfaces, not concrete types (`Reader`, `Writer`, `UserRepository`)
- Use apperror for all errors (validation/404/401/409/500)
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
    t.Run("CreateUser - Success", func(t *testing.T) {
        // Arrange: setup and expectations
        mockRepo := new(mocks.MockUserRepository)
        service := services.NewUserService(mockRepo)
        
        // Act: execute function
        result, err := service.CreateUser(&User{Email: "test@example.com"})
        
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
- Mock external dependencies (repositories, services)
- In-memory SQLite for repository unit tests
- Both success and failure test cases
- Never use `t.Errorf` (use testify)

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
- Create test files without `t.Run()`
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
5. **Server builds:** `go build -o /tmp/server ./cmd/server/main.go`
6. **Database migrations work:** `go run cmd/seeder/seeder.go`

## Configuration & Environment

Environment variables (from `.env` file, see `internal/configs/env.go`):
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `JWT_SECRET` - Secret key for JWT token signing
- `JWT_EXPIRY` - Token expiration in seconds (default: 900)
- `REFRESH_TOKEN_EXPIRY` - Refresh token expiration in seconds (default: 604800)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` - Email configuration
- `APP_PORT` - Server port (default: 8080)

For local development, see `.env.example` in root directory.

## When Adding New Features

1. **Start with the model** - Define domain model with validation tags
2. **Add repository layer** - Define interface, implement GORM ops, write unit tests
3. **Add service layer** - Business logic, validate inputs, orchestrate repos
4. **Add handler layer** - Accept requests, call service, return responses
5. **Add routes** - Register handler in internal/routes/routes.go
6. **Run all tests** - `go test ./... -v --cover` must pass with coverage targets
7. **Update docs** - If introducing new patterns, update DEVELOPMENT.md and TESTING.md

## Debugging Tips

- Use `go test -run TestName -v` for specific tests
- Use `go test -race` for race conditions
- Check `internal/configs/database_test.go` for test database setup
- Check `tests/mocks/` for mock examples
- View swagger at `/docs` when server is running

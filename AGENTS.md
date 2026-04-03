# AGENTS.md - Agent Coding Guidelines for golang-cms

This document provides guidelines for agentic coding agents operating in this repository.

## Project Overview

Go 1.25+ CMS REST API with JWT auth, MFA (TOTP), and clean architecture.
- **Framework:** Gin + GORM
- **Database:** MySQL 8.0+
- **Testing:** Testify with mocks

## Build, Lint & Test Commands

### Prerequisites
- Go 1.25+
- MySQL 8.0+
- Docker (for local development)

### Build Commands
```bash
# Build server binary
make build                          # Output: ./bin/server

# Or directly
go build -o bin/server ./cmd/server/main.go
```

### Test Commands
```bash
# Run all unit tests
go test ./... -v --cover

# Run single test (most common)
go test ./internal/handlers -run TestUserHandler -v

# Run specific test in a package
go test ./internal/services -run "TestUserService/CreateUser" -v

# Run tests with race detection
go test ./... -race

# Run E2E tests
go test ./tests/e2e/... -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run via Make (recommended)
make test                    # Unit tests with gotestsum
make test-coverage           # Tests with coverage report
make test-e2e                # E2E tests
```

### Lint & Format Commands
```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Full lint (requires golangci-lint)
golangci-lint run ./...

# Via Make
make fmt
make vet
make lint
```

### Pre-push Checklist
```bash
make pre-push    # Runs: fmt vet lint test
```

## Code Style Guidelines

### Architecture Layers
- **Handlers:** Parse requests, call services, return responses
- **Services:** Business logic, validation, orchestrate repositories
- **Repositories:** DB operations, implement interfaces
- **Models:** Domain objects with GORM/JSON tags

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case | `user_handler.go`, `user_service.go` |
| Interfaces | PascalCase with -er suffix | `UserHandler`, `UserService`, `UserRepository` |
| Functions | PascalCase | `GetUserByID`, `CreateUser` |
| Variables | camelCase | `userService`, `httpStatusCode` |
| Constants | PascalCase | `StatusCodeOK`, `ErrorCodeNotFound` |
| Packages | snake_case | `internal/handlers`, `pkg/logger` |
| JSON tags | snake_case | `user_id`, `created_at`, `updated_at` |
| DB columns | snake_case | `user_id`, `created_at` |

### Import Organization
Group imports in this order (blank line between groups):
1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/require"

    "github.com/vfa-khuongdv/golang-cms/internal/services"
    "github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)
```

### Error Handling

**Always use `apperror` package** for all errors:
```go
// Validation error
apperror.New(http.StatusBadRequest, validationError.Code, validationError.Message)

// Not found
apperror.New(http.StatusNotFound, 404, "User not found")

// Unauthorized
apperror.New(http.StatusUnauthorized, 401, "Invalid credentials")

// Internal server error
apperror.Wrap(http.StatusInternalServerError, 500, "Failed to create user", err)
```

**Never:**
- Ignore errors silently with `_`
- Use `panic()`
- Return generic `errors.New("error")`

### Dependency Injection
- Depend on interfaces, not concrete types
- Use constructor functions (e.g., `NewUserService`)

```go
type userServiceImpl struct {
    userRepo    repositories.UserRepository
    bcryptSvc   services.BcryptService
    jwtSvc      services.JWTService
}

func NewUserService(
    userRepo repositories.UserRepository,
    bcryptSvc services.BcryptService,
    jwtSvc services.JWTService,
) services.UserService {
    return &userServiceImpl{
        userRepo:  userRepo,
        bcryptSvc: bcryptSvc,
        jwtSvc:    jwtSvc,
    }
}
```

### Function Naming Patterns
- `Get...` - Retrieve single item
- `List...` - Retrieve multiple items
- `Create...` - Create new item
- `Update...` - Update existing item
- `Delete...` - Delete item
- `Is...` / `Has...` - Boolean checks

### Interface Design
Keep interfaces small (1-3 methods preferred):
```go
type UserRepository interface {
    Create(user *models.User) error
    GetByID(id string) (*models.User, error)
    GetByEmail(email string) (*models.User, error)
    Update(user *models.User) error
    Delete(id string) error
}
```

### Testing Standards

**Framework:** Testify with AAA pattern (Arrange-Act-Assert)
```go
func TestUserService_CreateUser(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        // Arrange
        mockRepo := new(mocks.MockUserRepository)
        mockRepo.On("Create", mock.Anything).Return(nil)
        service := services.NewUserService(mockRepo)

        // Act
        result, err := service.CreateUser(&models.User{Email: "test@example.com"})

        // Assert
        require.NoError(t, err)
        assert.NotNil(t, result)
        mockRepo.AssertExpectations(t)
    })
}
```

**Key rules:**
- Use `require` for hard assertions (fail fast)
- Use `assert` for soft assertions (collect all failures)
- Mock external dependencies (repositories, services)
- Test both success and failure cases
- Group tests with `t.Run()`
- Never use `t.Errorf` (use testify instead)

### Coverage Targets
- Handlers: 95%
- Services: 85%
- Repositories: 90%
- Middlewares: 85%
- Utils: 80%

### Configuration
- Use environment variables via `.env` file
- Access via `configs.LoadEnv()` or `utils.GetEnv()`
- Never hardcode config values

### Security Guidelines
- Never log sensitive information (passwords, tokens)
- Never store passwords in plain text (use bcrypt)
- Use JWT for authentication
- Validate all inputs

## Running the Application

```bash
# Start database
docker-compose up -d mysql

# Run migrations
go run cmd/seeder/seeder.go

# Start server
go run cmd/server/main.go
# Or: make dev (with hot reload via Air)
```

## When Adding New Features

1. **Model:** Define domain model with validation tags
2. **Repository:** Define interface, implement GORM ops, write unit tests
3. **Service:** Business logic, validate inputs, orchestrate repos
4. **Handler:** Accept requests, call service, return responses
5. **Routes:** Register handler in `internal/routes/routes.go`
6. **Tests:** Ensure all tests pass with coverage targets
7. **Validate:** Run `make pre-push`

## Key Files & Locations

| Path | Purpose |
|------|---------|
| `cmd/server/main.go` | Application entry point |
| `cmd/seeder/seeder.go` | Database seeding |
| `internal/handlers/` | HTTP handlers |
| `internal/services/` | Business logic |
| `internal/repositories/` | Database operations |
| `internal/models/` | Domain models |
| `internal/shared/dto/` | Data transfer objects |
| `internal/shared/utils/` | Utility functions |
| `pkg/apperror/` | Custom error handling |
| `pkg/logger/` | Logging utilities |
| `pkg/mailer/` | Email functionality |
| `tests/mocks/` | Test mocks |
| `tests/e2e/` | End-to-end tests |

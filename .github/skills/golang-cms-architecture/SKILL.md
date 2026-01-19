---
name: golang-cms-architecture
description: Complete development guidelines for golang-cms, a Go-based CMS REST API with clean architecture (handlers → services → repositories → models). Use when implementing features, writing tests, or modifying the golang-cms codebase. Covers authentication, MFA, JWT tokens, testing patterns with testify, error handling, and naming conventions.
license: MIT
metadata:
  author: golang-cms
  version: "1.0"
---

# Golang CMS Architecture & Development Guide

## Project Overview

This project targets Go 1.22+ (minimum supported version).
- User authentication with JWT tokens and refresh tokens
- Clean architecture: handlers → services → repositories → models
- Comprehensive testing with testify (assert, require, mock)
- Standardized error handling via apperror package
- Framework: Gin, Database: GORM with MySQL

## Project Structure

```
├── cmd/
│   ├── server/main.go          # Application entry point
│   └── seeder/seeder.go        # Database seeding utility
├── internal/
│   ├── handlers/               # HTTP handlers (Gin)
│   ├── services/               # Business logic layer
│   ├── repositories/           # Data access layer (GORM)
│   ├── models/                 # Domain models
│   ├── middlewares/            # HTTP middlewares (auth, CORS, logging)
│   ├── configs/                # Configuration and environment loading
│   ├── constants/              # Application constants
│   ├── routes/                 # Route definitions
│   ├── utils/                  # Utility functions
│   └── database/               # Migrations and seeders
├── pkg/
│   ├── apperror/               # Error codes and error types
│   ├── logger/                 # Logging utility
│   ├── mailer/                 # Email/SMTP service
│   └── migrator/               # Database migration runner
└── tests/mocks/                # Mock implementations for testing
```

## Core Architecture Principles

### Layer Responsibilities

**Handlers** (internal/handlers/):
- Parse HTTP requests
- Call services
- Return HTTP responses
- Minimal business logic
- Handle request validation format (not business validation)

**Services** (internal/services/):
- Business logic
- Input validation
- Orchestrate repositories
- Return errors via apperror package

**Repositories** (internal/repositories/):
- Database operations using GORM
- Return domain entities
- Implement interfaces

**Models** (internal/models/):
- Domain objects
- GORM tags for database mapping
- JSON tags in snake_case for API responses
- Validation tags

### Dependency Injection Pattern

All layers depend on abstractions (interfaces), never concrete implementations:

```go
// Good - depends on interface
type UserService struct {
    repo UserRepository  // Interface, not concrete type
}

func NewUserService(repo UserRepository) UserService {
    return &userServiceImpl{repo: repo}
}

// Bad - directly depends on concrete type
type userServiceImpl struct {
    repo userRepositoryImpl  // Concrete implementation, not reusable
}
```

Keep interfaces small and focused (Interface Segregation Principle):
- Prefer 1-3 methods per interface
- Easier to mock and test
- Follow idiomatic Go naming: `Reader`, `Writer`, `UserRepository` (no `I` prefix)

## Naming Conventions

### Packages & Files
- Use lowercase, single-word names: `handlers`, `services`, `repositories`
- Test files: `user_service_test.go`, `auth_handler_test.go`, `*_integration_test.go`

### Functions & Methods
- Use verb-based names for actions
- `Get` for retrieval: `GetUser(id)`, `GetUserByEmail(email)`
- `Create`/`Update`/`Delete` for modifications: `CreateUser(req)`, `DeleteUser(id)`
- `Is`/`Has` for booleans: `IsValidEmail(email)`, `HasPermission(user, action)`
- `New` for constructors: `NewUserService(repo)`

### Constants
- All uppercase with underscores: `MAX_RETRY_COUNT`, `DEFAULT_TIMEOUT`, `JWT_SECRET`
- Group related constants together

### JSON Tags (API Responses)
- Use snake_case (REST convention): `created_at`, `user_id`, `is_active`, `email`
- Example:
  ```go
  type User struct {
      ID        uint      `json:"id"`
      Email     string    `json:"email"`
      CreatedAt time.Time `json:"created_at"`
      UpdatedAt time.Time `json:"updated_at"`
  }
  ```

### Test Functions & Subtests
- Format: `TestFunctionName` or grouped under parent: `TestUserService`
- Use `t.Run()` for organized subtests with descriptive names
- Example:
  ```go
  func TestUserService(t *testing.T) {
      t.Run("CreateUser - Success", func(t *testing.T) { ... })
      t.Run("CreateUser - Validation Error", func(t *testing.T) { ... })
  }
  ```

## Error Handling

Always use the `apperror` package for standardized errors:

```go
import "github.com/vfa-khuongdv/golang-cms/pkg/apperror"

// Validation errors (HTTP 400)
return nil, apperror.NewValidationError("Email is required")

// Not found errors (HTTP 404)
return nil, apperror.NewNotFoundError("User not found")

// Authentication errors (HTTP 401)
return nil, apperror.NewUnauthorizedError("Invalid credentials")

// Conflict errors (HTTP 409)
return nil, apperror.NewConflictError("Email already exists")

// Server errors (HTTP 500) - wrap with context
return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternal, "Failed to create user", originalErr)
```

**HTTP Status Mapping:**
- 400: Validation errors
- 401: Authentication errors
- 403: Forbidden/Authorization errors
- 404: Not found errors
- 409: Conflict errors
- 500: Server errors

Never ignore errors silently. Always handle explicitly.

## Testing Standards

### Test Structure (AAA Pattern)

Every test follows Arrange-Act-Assert:

```go
func TestUserService(t *testing.T) {
    t.Run("CreateUser - Success", func(t *testing.T) {
        // ARRANGE - Setup test data and expectations
        mockRepo := new(mocks.MockUserRepository)
        service := services.NewUserService(mockRepo)
        user := &User{Email: "test@example.com", Name: "Test User"}
        
        mockRepo.On("Create", mock.MatchedBy(func(u *User) bool {
            return u.Email == user.Email
        })).Return(nil).Once()

        // ACT - Execute the function being tested
        result, err := service.CreateUser(user)

        // ASSERT - Verify the results
        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, user.Email, result.Email)
        mockRepo.AssertExpectations(t)
    })
}
```

### Testify Packages

**assert** - Soft assertions (continue on failure):
```go
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Nil(t, value)
assert.NotNil(t, value)
```

**require** - Hard assertions (stop on failure) for critical checks:
```go
require.NoError(t, err)
require.NotNil(t, result)
```

**mock** - Object mocking for dependencies:
```go
mockRepo.On("Method", arg1, arg2).Return(result, nil)
mockRepo.AssertExpectations(t)
```

### Test Coverage Requirements by Layer

- **Handlers**: 95%+ coverage
- **Services**: 85%+ coverage
- **Repositories**: 90%+ coverage
- **Middlewares**: 85%+ coverage
- **Utils**: 80%+ coverage

Run coverage: `go test ./... --cover`

### Repository Testing

Use in-memory SQLite for unit tests:
```go
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
// Create test fixtures
// Run assertions
// Database is cleaned up automatically
```

### Handler Testing

```go
func TestUserHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    t.Run("CreateUser - Success", func(t *testing.T) {
        // Arrange
        mockService := new(mocks.MockUserService)
        user := &User{ID: 1, Email: "test@example.com"}
        mockService.On("CreateUser", mock.Anything).Return(user, nil)
        
        handler := handlers.NewUserHandler(mockService)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        
        body, _ := json.Marshal(map[string]any{"email": "test@example.com"})
        c.Request, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
        c.Request.Header.Set("Content-Type", "application/json")
        
        // Act
        handler.CreateUser(c)
        
        // Assert
        require.Equal(t, http.StatusCreated, w.Code)
        mockService.AssertExpectations(t)
    })
}
```

### Mocking Guidelines

- Mock external dependencies only (repositories, services)
- Don't mock internal dependencies or return types
- Always call `AssertExpectations(t)` after using a mock
- Use `mock.MatchedBy()` for complex matching logic
- Use `.Once()` or `.Times(n)` to specify call count expectations

### Key Testing Rules

✓ **DO:**
- Use testify/assert and testify/require
- Follow AAA pattern
- Group related tests with `t.Run()`
- Use `require` for critical checks, `assert` for value checks
- Mock all external dependencies
- Write both success and failure test cases
- Clean up resources in teardowns

✗ **DON'T:**
- Ignore errors (never `_ = err`)
- Write tests that depend on each other
- Skip error handling tests
- Test multiple things in one test
- Use time-based assertions (flaky)
- Mix unit and integration tests in same function

## Adding New Features

Follow this step-by-step workflow:

1. **Start with the model** (internal/models/)
   - Define domain model with GORM and JSON tags
   - Add validation tags if needed

2. **Add repository layer** (internal/repositories/)
   - Define interface first (UserRepository, not IUserRepository)
   - Implement GORM operations
   - Write unit tests with in-memory SQLite
   - Target 90%+ coverage

3. **Add service layer** (internal/services/)
   - Implement business logic
   - Validate inputs
   - Orchestrate repositories
   - Write tests with mocked repositories
   - Target 85%+ coverage

4. **Add handler layer** (internal/handlers/)
   - Accept HTTP requests (Gin)
   - Validate request format
   - Call service
   - Return HTTP responses
   - Write handler tests
   - Target 95%+ coverage

5. **Add routes** (internal/routes/)
   - Register handler in routes.go

6. **Run all tests**
   ```bash
   go test ./... -v --cover
   ```
   All tests must pass and meet coverage targets.

## Build & Validation Commands

```bash
# Download dependencies
go mod download

# Run all tests with coverage
go test ./... -v --cover

# Run specific package tests
go test ./internal/handlers -v

# Run specific test
go test ./internal/handlers -run TestUserHandler -v

# Race condition detection
go test ./... -race

# Format check
go fmt ./...

# Vet analysis
go vet ./...

# Build server
make build

# Start development environment
docker-compose up -d mysql
go run cmd/seeder/seeder.go
go run cmd/server/main.go
```

## Important Implementation Rules

### ALWAYS DO:
- Use dependency injection for testability
- Handle all errors explicitly
- Use apperror package for all application errors
- Write tests immediately after writing code
- Use meaningful, descriptive names
- Add comments explaining "why", not "what"
- Keep interfaces small (1-3 methods)
- Use snake_case for JSON tags
- Mock external dependencies
- Follow clean code principles

### NEVER DO:
- Ignore errors silently
- Use global variables
- Mix concerns between layers
- Hardcode configuration values
- Store passwords in plain text
- Log sensitive information
- Mix unit and integration tests
- Create test files without grouped subtests
- Bypass AuthMiddleware on protected routes
- Have complex setup in tests

## Authentication & Authorization

**JWT Tokens:**
- Generated by JWTService
- Configurable expiration (default: 900s = 15 minutes)
- Validated by AuthMiddleware on protected routes
- Stored in Authorization header: `Bearer <token>`

**Refresh Tokens:**
- Generated during login
- Stored in database via RefreshTokenRepository
- Configurable expiration (default: 604800s = 7 days)
- Used to obtain new JWT tokens

**Middleware:**
- AuthMiddleware validates JWT on protected routes
- CorsMiddleware handles cross-origin requests
- LogMiddleware logs all requests
- EmptyBodyMiddleware handles empty request bodies

## Environment Configuration

Environment variables (from `.env` or passed to application):

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=golang_cms

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=900

# Refresh Token
REFRESH_TOKEN_EXPIRY=604800

# Email (SMTP)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASSWORD=password

# Server
APP_PORT=3000
```

## Useful References

See the project's `DEVELOPMENT.md` and `TESTING.md` files for detailed guidelines on:
- Specific naming conventions
- Code style guidelines
- Testing patterns and examples
- Database setup
- Migration procedures

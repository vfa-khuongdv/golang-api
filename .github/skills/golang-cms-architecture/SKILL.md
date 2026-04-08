---
name: golang-cms-architecture
description: Complete development guidelines for golang-cms, a Go-based CMS REST API with clean architecture (handlers → services → repositories → models). Use when implementing features, writing tests, or modifying the golang-cms codebase. Covers authentication, JWT tokens, testing patterns with testify, error handling, and naming conventions.
license: MIT
metadata:
  author: golang-cms
  version: "1.0"
---

# Golang CMS Architecture & Development Guide

> **Additional Resources**: See `references/` folder for templates, cheatsheet, and workflow guides.

## Project Overview

This project targets Go 1.22+ (minimum supported version).
- User authentication with JWT tokens and refresh tokens
- Clean architecture: handlers → services → repositories → models
- Comprehensive testing with testify (assert, require, mock)
- Standardized error handling via apperror package
- Framework: Gin, Database: GORM with MySQL

## Project Structure

```
├── cmd/                          # Command-line applications
│   ├── server/                   # Main application entry point
│   │   └── main.go
│   └── seeder/                   # Database seeder
│       └── seeder.go
├── internal/                     # Private application code
│   ├── configs/                  # Configuration management
│   ├── database/                 # Database setup
│   │   ├── migrations/           # Migration files
│   │   └── seeders/              # Seeder implementations
│   ├── handlers/                 # HTTP handlers/controllers
│   ├── middlewares/              # HTTP middlewares
│   ├── models/                   # Data models
│   ├── repositories/             # Data access layer
│   ├── routes/                   # Route definitions
│   ├── services/                 # Business logic layer    
│   └── shared/                   # Shared utilities and helpers
│       └── constants/            # Application constants
│       └── dto/                  # Data transfer objects for shared use
│       └── utils/                # Utility functions  for shared use
├── pkg/                          # Public packages
│   ├── apperror/                 # Custom error handling
│   ├── logger/                   # Logging utilities
│   ├── mailer/                   # Email sending utilities
│   └── migrator/                 # Database migration utilities
├── tests/                        # Test utilities and mocks
│   └── mocks/                    # Mock implementations
├── docs/                         # Documentation and API specs
└── Makefile                      # Build and development commands
```

## Core Architecture Principles

> See `references/cheatsheet.md` for quick reference on patterns and logging.

### Layer Responsibilities

| Layer | Responsibility |
|-------|----------------|
| **Handlers** | Parse requests, call services, return responses |
| **Services** | Business logic, validation, orchestrate repos |
| **Repositories** | DB operations using GORM, implement interfaces |
| **Models** | Domain objects with GORM/JSON tags |

All service and repository methods must accept `context.Context` as the first parameter.

### Dependency Injection

Always depend on interfaces, not concrete types:
```go
type UserService interface { /* ... */ }
type userServiceImpl struct { repo UserRepository }

func NewUserService(repo UserRepository) UserService {
    return &userServiceImpl{repo: repo}
}
```

Keep interfaces small (1-3 methods).

### Context Propagation

Handlers pass `ctx.Request.Context()` to services. Services pass to repositories. Use `db.WithContext(ctx)` before GORM operations.

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

// Server errors (HTTP 500)
return nil, apperror.NewInternalServerError("Failed to create user: %w", err)

```

**HTTP Status Mapping:**
- 400: Validation or Bad Request errors
- 401: Authentication errors
- 403: Forbidden/Authorization errors
- 404: Not found errors
- 500: Server errors

Never ignore errors silently. Always handle explicitly.

## Response API
- User utils.ResponseWithOK, ResponseWithError for consistent API responses
- Example:

```go
utils.ResponseWithOK(c, http.StatusCreated, "User created successfully", user)
utils.ResponseWithError(c, err)
```

## Testing Standards

Follow AAA pattern with testify:
```go
func TestUserService(t *testing.T) {
    t.Run("CreateUser - Success", func(t *testing.T) {
        // ARRANGE
        mockRepo := new(mocks.MockUserRepository)
        mockRepo.On("Create", mock.Anything).Return(&models.User{}, nil)
        
        // ACT
        result, err := services.NewUserService(mockRepo).CreateUser(ctx, input)
        
        // ASSERT
        require.NoError(t, err)
        assert.NotNil(t, result)
    })
}
```

**Coverage targets:** Handlers 95%, Services 85%, Repos 90%, Middlewares 85%, Utils 80%

> See `references/templates.md` for detailed test examples.

## Adding New Features

1. **Model** → Define domain model with GORM/JSON tags
2. **Repository** → Define interface, implement GORM ops, write tests (90%+)
3. **Service** → Business logic, validation, orchestrate repos, write tests (85%+)
4. **Handler** → Parse requests, call service, return responses, write tests (95%+)
5. **Routes** → Register handler in routes.go
6. **Validate** → Run tests

> See `references/workflow.md` for detailed development workflow.

## Build Commands

```bash
make build              # Build binary
make dev                # Start with hot reload
```

> See `references/commands.md` for full command reference.

## Important Implementation Rules

**DO:** Dependency injection, explicit errors, apperror package, context.Context, logger.WithContext(ctx), tests immediately, snake_case JSON tags

**DON'T:** Ignore errors, global variables, mix concerns, hardcode config, plain text passwords, log sensitive info, complex test setup

## Logging

Use `logger.WithContext(ctx)` for request-scoped logging (auto-includes request_id).
Use plain `logger.Infof()` for startup/seeders.

> See `references/cheatsheet.md` for full logging patterns.

## Authentication

- **JWT:** 15 min default, validated by AuthMiddleware
- **Refresh Token:** 7 days default, stored in database
- **Middleware:** Auth, CORS, Log, EmptyBody

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

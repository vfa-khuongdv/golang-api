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

This project targets Go 1.27+ (see `go.mod`).
- User authentication with JWT access tokens (1 hour) and rotating refresh tokens (30 days)
- Account lockout, per-IP rate limiting on public auth routes, and password reset by email
- Key/value `settings` table for mail and frontend settings (secret values encrypted with AES-256-GCM)
- Clean architecture: handlers → services → repositories → models
- Comprehensive testing with testify (assert, require, mock)
- Standardized error handling via apperror package
- Framework: Gin, Database: GORM with MySQL

## Project Structure

```
├── cmd/                          # Command-line applications
│   ├── encrypt-setting/          # Encrypts secret settings (make encrypt-setting)
│   │   └── main.go
│   ├── server/                   # Main application entry point
│   │   └── main.go
│   └── seeder/                   # Database seeder
│       └── seeder.go
├── internal/                     # Private application code
│   ├── configs/                  # Configuration management
│   ├── database/                 # Database setup
│   │   ├── migrations/           # SQL migrations (golang-migrate)
│   │   └── seeders/              # Seeder implementations
│   ├── handlers/                 # HTTP handlers/controllers
│   ├── middlewares/              # HTTP middlewares
│   ├── models/                   # Data models
│   ├── repositories/             # Data access layer
│   ├── routes/                   # Route definitions
│   ├── services/                 # Business logic layer    
│   └── shared/                   # Shared utilities and helpers
│       ├── constants/            # Application constants
│       ├── dto/                  # Data transfer objects for shared use
│       └── utils/                # Utility functions  for shared use
├── pkg/                          # Public packages
│   ├── apperror/                 # Custom error handling
│   ├── logger/                   # Logging utilities
│   ├── mailer/                   # Email sending utilities
│   └── migrator/                 # Database migration utilities
├── tests/                        # Shared test support (unit tests sit next to the code)
│   ├── e2e/                      # End-to-end tests
│   └── mocks/                    # Mock implementations
├── docs/                         # Swagger spec and logging standards
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
type userServiceImpl struct {
    repo          repositories.UserRepository
    mailerService MailerService
}

func NewUserService(repo repositories.UserRepository, mailerService MailerService) UserService {
    return &userServiceImpl{repo: repo, mailerService: mailerService}
}
```

Keep interfaces small (1-3 methods).

### Context Propagation

Handlers pass `ctx.Request.Context()` to services. Services pass to repositories. Use `db.WithContext(ctx)` before GORM operations.

## Naming Conventions

### Packages & Files
- Use lowercase, single-word names: `handlers`, `services`, `repositories`
- Test files sit next to the code: `user_service_test.go`, `auth_handler_test.go`; use `*_internal_test.go` for tests that need unexported identifiers

### Functions & Methods
- Use verb-based names for actions
- `Get` for retrieval: `GetUser(id)`, `GetUserByEmail(email)`
- `Create`/`Update`/`Delete` for modifications: `CreateUser(req)`, `DeleteUser(id)`
- `Is`/`Has` for booleans: `IsValidEmail(email)`, `HasPermission(user, action)`
- `New` for constructors: `NewUserService(repo)`

### Constants
- Idiomatic Go CamelCase for new constants: `MaxFailedAttempts`, `SettingMailHost`
- A few older constants use upper snake case (`LIMIT`, `MAX_BODY_SIZE`); leave them as they are
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

// Validation errors (HTTP 400): normally produced in handlers from binding errors
utils.RespondWithError(c, utils.TranslateValidationErrors(err, input))
// or built manually:
// apperror.NewValidationError("Validation failed", []apperror.FieldError{{Field: "email", Message: "email is required"}})

// Bad request (HTTP 400)
return nil, apperror.NewBadRequestError("Invalid input")

// Not found errors (HTTP 404)
return nil, apperror.NewNotFoundError("User not found")

// Authentication errors (HTTP 401)
return nil, apperror.NewUnauthorizedError("Invalid credentials")

// Conflict errors (HTTP 409)
return nil, apperror.NewConflictError("Email already exists")

// Server errors (HTTP 500)
return nil, apperror.NewInternalServerError("Failed to create user")
// or keep the cause: apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to create user", err)

```

**HTTP Status Mapping:**
- 400: Validation, bad request, and password errors
- 401: Authentication errors
- 403: Forbidden/Authorization errors
- 404: Not found errors
- 409: Conflict errors
- 429: Account locked / too many requests
- 500: Server errors

Error codes live in `pkg/apperror/codes.go`. Error bodies are `{"code": ..., "message": ...}`; validation errors add a `fields` array.

Never ignore errors silently. Always handle explicitly.

## Response API
- Use `utils.RespondWithOK(ctx, statusCode, body)` and `utils.RespondWithError(ctx, err)` for consistent API responses
- `RespondWithOK` writes `body` as the JSON response as-is (no envelope)
- Example:

```go
utils.RespondWithOK(c, http.StatusOK, gin.H{"message": "Update profile successfully"})
utils.RespondWithError(c, err)
```

## Testing Standards

Follow AAA pattern with testify:
```go
func TestUserService(t *testing.T) {
    t.Run("CreateUser - Success", func(t *testing.T) {
        // ARRANGE
        mockRepo := new(mocks.MockUserRepository)
        mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{ID: 1}, nil)

        // ACT
        result, err := services.NewUserService(mockRepo, new(mocks.MockMailerService)).GetProfile(ctx, 1)
        
        // ASSERT
        require.NoError(t, err)
        assert.NotNil(t, result)
    })
}
```

**Coverage targets:** Handlers 95%, Services 85%, Repos 90%, Middlewares 85%, Utils 80% (CI fails below 70% total)

> See `references/templates.md` for detailed test examples.

## Adding New Features

1. **Model** → Define domain model with GORM/JSON tags
2. **Repository** → Define interface, implement GORM ops, write tests (90%+)
3. **Service** → Business logic, validation, orchestrate repos, write tests (85%+)
4. **Handler** → Parse requests, call service, return responses, write tests (95%+)
5. **Routes** → Register handler in routes.go
6. **Migration** → Add `*.up.sql` / `*.down.sql` files in `internal/database/migrations`
7. **Docs** → Update `docs/swagger.json` and the README endpoint list
8. **Validate** → Run tests

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
Use `logger.WithEvent(ctx, logger.EventLoginFailed)` for business events (adds the `event` field).
Use plain `logger.Infof()` for startup/seeders.
See `docs/logging-standards.md` for fields, events, and masking.

> See `references/cheatsheet.md` for full logging patterns.

## Authentication

- **JWT:** access token valid for 1 hour, `access` scope, validated by `AuthMiddleware`
- **Refresh Token:** random 60-character token valid for 30 days, stored in the database, rotated on refresh, deleted on logout
- **Lockout:** 5 failed logins lock the account for 15 minutes
- **Rate limit:** 10 requests/minute per IP on login, refresh-token, forgot-password, and reset-password
- **Middleware:** RequestID, CORS, Log, Recovery (global); RateLimiter (public routes); Auth (authenticated routes). Empty bodies are rejected by `utils.TranslateValidationErrors`

## Environment Configuration

Environment variables (from `.env` or passed to application):

```bash
# Database
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=password
DB_DATABASE=golang_cms

# JWT (at least 32 characters)
JWT_KEY=your-32-character-secret-key-here

# Settings encryption (at least 32 characters)
SETTINGS_ENCRYPTION_KEY=your-32-character-encryption-key-here

# Server
PORT=3000
RUN_MIGRATE=true
```

`DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`, `JWT_KEY`, and `SETTINGS_ENCRYPTION_KEY` are required. See `.env.example` and the README for optional variables.

Mail (`mail_*`) and frontend (`frontend_url`) settings are not environment variables; they live in the `settings` table (key/value) and are seeded by migrations. `mail_password` is stored encrypted with `SETTINGS_ENCRYPTION_KEY` (generate the value with `make encrypt-setting`).

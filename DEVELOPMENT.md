# Development Guidelines

This document outlines the coding standards, project structure, and best practices for the Golang CMS project. For setup instructions and the API overview see [README.md](README.md); for testing details see [TESTING.md](TESTING.md).

## Table of Contents

1. [Project Architecture](#project-architecture)
2. [Code Organization](#code-organization)
3. [Naming Conventions](#naming-conventions)
4. [Interface Design](#interface-design)
5. [Error Handling](#error-handling)
6. [Testing Guidelines](#testing-guidelines)
7. [Git Workflow](#git-workflow)
8. [Documentation](#documentation)
9. [Performance Guidelines](#performance-guidelines)
10. [Security Guidelines](#security-guidelines)
11. [Environment Variables](#environment-variables)
12. [Deployment](#deployment)
13. [Tools and Dependencies](#tools-and-dependencies)
14. [Common Mistakes to Avoid](#common-mistakes-to-avoid)
15. [Code Review Checklist](#code-review-checklist)

---

## Project Architecture

### Clean Architecture Pattern

This project follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP/REST Layer                      │
│                    (handlers, routes)                   │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    Service Layer                        │
│              (business logic, validation)               │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                  Repository Layer                       │
│            (database access, queries)                   │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                   Data Layer                            │
│                    (Database)                           │
└─────────────────────────────────────────────────────────┘
```

Dependencies are wired by hand in `internal/routes/routes.go` (`SetupRouter`): repositories → services → handlers.

### Benefits

- **Testability**: Each layer can be tested independently with mocks
- **Maintainability**: Changes in one layer don't affect others
- **Scalability**: Easy to add new features without affecting existing code
- **Flexibility**: Database or HTTP framework changes don't require major refactoring

---

## Code Organization

### Directory Structure

```
project/
├── cmd/                          # Command-line applications
│   ├── encrypt-setting/          # Encrypts secret setting values (make encrypt-setting)
│   │   └── main.go
│   ├── server/                   # Main application entry point
│   │   └── main.go
│   └── seeder/                   # Database seeder
│       └── seeder.go
├── internal/                     # Private application code
│   ├── configs/                  # Environment configuration and DB connection
│   ├── database/                 # Database setup
│   │   ├── migrations/           # SQL migrations (golang-migrate)
│   │   └── seeders/              # Seeder implementations
│   ├── handlers/                 # HTTP handlers/controllers
│   ├── middlewares/              # HTTP middlewares
│   ├── models/                   # Data models
│   ├── repositories/             # Data access layer
│   ├── routes/                   # Router setup and dependency wiring
│   ├── services/                 # Business logic layer
│   └── shared/                   # Shared utilities and helpers
│       ├── constants/            # Application constants
│       ├── dto/                  # Request/response data transfer objects
│       └── utils/                # Utility functions for shared use
├── pkg/                          # Reusable packages
│   ├── apperror/                 # Custom error handling
│   ├── logger/                   # Logging utilities
│   ├── mailer/                   # SMTP sender and embedded email templates
│   └── migrator/                 # Database migration utilities
├── tests/                        # Shared test support
│   ├── e2e/                      # End-to-end tests
│   └── mocks/                    # Mock implementations
├── docs/                         # Swagger spec and logging standards
└── Makefile                      # Build and development commands
```

Unit tests are not kept in `tests/`; they sit next to the code (`internal/**/*_test.go`, `pkg/**/*_test.go`).

### Responsibilities by Layer

#### **Models (internal/models/)**
- Define data structures
- Represent database tables (GORM tags) and JSON serialization (`json` tags)
- Hide internal fields from API responses with `json:"-"` (e.g. `Password`, `ResetToken`)
- Keep models simple - no business logic

#### **DTOs (internal/shared/dto/)**
- Define request inputs and response bodies
- Carry the `binding` validation tags (`required`, `email`, `password_complexity`, ...)

#### **Repositories (internal/repositories/)**
- Handle all database operations
- Implement CRUD operations
- Return domain entities
- No business logic - only data access
- Implement interfaces for mockability

#### **Services (internal/services/)**
- Contain business logic
- Enforce business rules (lockout, token expiry, password rules)
- Orchestrate between repositories
- Handle errors appropriately
- Return clean data to handlers

#### **Handlers (internal/handlers/)**
- Bind and validate HTTP requests
- Call appropriate services
- Return HTTP responses
- Handle response formatting

#### **Middlewares (internal/middlewares/)**
- Request ID (`X-Request-ID`)
- Authentication (JWT access tokens)
- CORS handling
- Request/response logging with sensitive data masking
- Rate limiting

Panic recovery is provided by Gin's `gin.Recovery()`. Empty request bodies are rejected centrally by `utils.TranslateValidationErrors`, not by a middleware.

---

## Naming Conventions

### Go Standards

- **Packages**: Lowercase, single word when possible
  ```go
  package handlers
  package services
  package repositories
  ```

- **Functions/Methods**: CamelCase, exported starts with uppercase
  ```go
  func GetProfile()        // Exported
  func getUserByEmail()    // Unexported
  ```

- **Constants**: Write new constants in idiomatic Go CamelCase (exported starts with uppercase). A few older constants use upper snake case (`LIMIT`, `DEFAULT_MAX_OPEN_CONNS`, `MAX_BODY_SIZE`); leave them as they are.
  ```go
  const (
      MaxFailedAttempts      = 5
      LockoutDurationMinutes = 15
      SettingMailHost        = "mail_host"
  )
  ```

- **Interfaces**: Descriptive names without `I` prefix (idiomatic Go)
  ```go
  type UserService interface {}      // Correct: no prefix
  type EmailSender interface {}      // Correct: descriptive
  type UserRepository interface {}   // Correct: descriptive
  ```

- **Variables**: CamelCase for exported, camelCase for unexported
  ```go
  var DB *gorm.DB                   // Exported
  var userRepository UserRepository // Unexported
  ```

### Naming Patterns by Layer

#### **Models**
- Use singular names for structs
- Use descriptive field names
- Use JSON tags for serialization in **snake_case** format for API responses
- Set the table name explicitly with `TableName()`

```go
type Setting struct {
    ID        uint      `gorm:"column:id;primaryKey" json:"id"`
    Key       string    `gorm:"column:key;type:varchar(100);not null;unique" json:"key"`
    Value     string    `gorm:"column:value;type:text;not null" json:"value"`
    CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Setting) TableName() string { return "settings" }
```

**Note:** All JSON field names in API responses must use snake_case (e.g., `created_at`, `user_id`, `expires_at`) for consistency with REST API conventions.

#### **Interfaces**
- Describe the contract clearly
- Use descriptive names without `I` prefix (idiomatic Go)
- Keep interfaces small and focused
- Repositories and services return interface types from their constructors; handlers return their concrete struct pointer
- Every service and repository method takes `context.Context` as its first parameter

```go
type SettingRepository interface {
    GetValues(ctx context.Context, keys ...string) (map[string]string, error)
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
    return &settingRepositoryImpl{db: db}
}
```

#### **Methods**
- Use verb-based names for actions
- Use `Get` for retrieval
- Use `Create`, `Update`, `Delete` for modifications
- Use `Is`, `Has` for boolean checks

```go
func (s *userServiceImpl) GetProfile(ctx context.Context, userID uint) (*models.User, error)
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error
func (s *userServiceImpl) ChangePassword(ctx context.Context, userID uint, input *dto.ChangePasswordInput) (*models.User, error)
```

#### **Test Functions**
- Use format: `TestFunctionName` or `TestFunctionName/subtest`
- Group related tests under parent test function
- Use descriptive subtest names

```go
func TestUserRepository(t *testing.T) {
    t.Run("GetAll - Success", func(t *testing.T) {
        // test code
    })
    t.Run("GetAll - Empty", func(t *testing.T) {
        // test code
    })
}
```

---

## Interface Design

### Principles

1. **Depend on Abstractions**
   ```go
   // Good - depends on interface
   type userServiceImpl struct {
       repo          repositories.UserRepository
       mailerService MailerService
   }

   // Bad - depends on concrete type
   type userServiceImpl struct {
       repo *userRepositoryImpl
   }
   ```

2. **Single Responsibility**
   ```go
   // Good - focused interface (see repositories.SettingRepository)
   type SettingRepository interface {
       GetValues(ctx context.Context, keys ...string) (map[string]string, error)
   }

   // Bad - too many responsibilities
   type Database interface {
       Query(sql string) Results
       Execute(sql string) error
       GetUser(id uint) (*User, error)
       CreateUser(user *User) error
   }
   ```

3. **Interface Segregation**
   ```go
   // Good - small, specific interface (see services.MailerService)
   type MailerService interface {
       SendMailForgotPassword(ctx context.Context, user *models.User) error
   }

   // Bad - large, monolithic interface
   type FileHandler interface {
       Read() ([]byte, error)
       Write([]byte) error
       Delete() error
       // ... many more methods
   }
   ```

   Some interfaces in the codebase are larger than ideal (`UserRepository`, `UserService`); prefer small interfaces for new code.

### Handler Pattern

Handlers are concrete structs that receive their service dependencies through the constructor. They are not hidden behind an interface; tests call the methods directly.

```go
type userHandlerImpl struct {
    userService   services.UserService
    mailerService services.MailerService
}

func NewUserHandler(
    userService services.UserService,
    mailerService services.MailerService,
) *userHandlerImpl {
    return &userHandlerImpl{
        userService:   userService,
        mailerService: mailerService,
    }
}
```

---

## Error Handling

### Custom Error Structure

Use the `apperror` package (`pkg/apperror`) for consistent error handling:

```go
type AppError struct {
    HttpStatusCode int    `json:"-"`       // HTTP status code
    Code           int    `json:"code"`    // Application error code (see codes.go)
    Message        string `json:"message"` // Client-facing message
    Err            error  `json:"-"`       // Underlying error (optional)
}
```

Error codes are grouped in `pkg/apperror/codes.go`: general (1000s), database (2000s), authentication (3000s), and common/cache (4000s).

### Error Creation

```go
// Create a new error with an explicit status and code
err := apperror.New(http.StatusTooManyRequests, apperror.ErrTooManyRequests, "Too many requests")

// Wrap an existing error
err := apperror.Wrap(
    http.StatusInternalServerError,
    apperror.ErrInternalServer,
    "Failed to create user",
    originalError,
)

// Factory helpers (each sets the HTTP status and code)
apperror.NewBadRequestError("Invalid input")              // 400
apperror.NewUnauthorizedError("Invalid credentials")      // 401
apperror.NewForbiddenError("Access denied")               // 403
apperror.NewNotFoundError("User not found")               // 404
apperror.NewConflictError("Email already exists")         // 409
apperror.NewInternalServerError("Server error")           // 500
apperror.NewInvalidPasswordError("Invalid credentials")   // 400
apperror.NewAccountLockedError("Account is locked")       // 429
apperror.NewDBQueryError("Failed to query users")         // 500
```

Validation failures use a separate type, `apperror.ValidationError`, which carries per-field messages:

```go
apperror.NewValidationError("Validation failed", []apperror.FieldError{
    {Field: "email", Message: "email is required"},
})
```

### Error Handling in Handlers

Bind JSON, translate binding errors into a `ValidationError`, call the service, and respond through the `utils` helpers. `RespondWithError` accepts any error: `ValidationError` → 400 with `fields`, `AppError` → its own status, anything else → 500 "Internal server error".

```go
func (handler *userHandlerImpl) ChangePassword(ctx *gin.Context) {
    userId, err := utils.GetUserIDFromContext(ctx)
    if err != nil {
        utils.RespondWithError(ctx, apperror.NewParseError("Invalid UserID"))
        return
    }

    var input dto.ChangePasswordInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        utils.RespondWithError(ctx, utils.TranslateValidationErrors(err, input))
        return
    }

    _, err = handler.userService.ChangePassword(ctx.Request.Context(), userId, &input)
    if err != nil {
        logger.WithEvent(ctx.Request.Context(), logger.EventPasswordChangeFailed).
            Errorf("Change password failed for user %d: %v", userId, err)
        utils.RespondWithError(ctx, err)
        return
    }

    utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
}
```

### Error Handling in Services

Services return `*apperror.AppError` values; they never expose raw database errors to the client.

```go
func (service *userServiceImpl) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
    user, err := service.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, apperror.NewNotFoundError("User not found")
    }
    return user, nil
}

func (service *userServiceImpl) UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error {
    user, err := service.repo.GetByID(ctx, userID)
    if err != nil {
        return apperror.NewNotFoundError("User not found")
    }
    // ... apply the fields that were provided ...
    if err := service.repo.Update(ctx, user); err != nil {
        logger.WithEvent(ctx, logger.EventProfileUpdateFailed).Errorf("Failed to update user profile: %v", err)
        return apperror.NewDBUpdateError("Failed to update profile")
    }
    return nil
}
```

---

## Testing Guidelines

### Test Organization

All tests should follow the **Testify** framework with proper grouping:

1. **Group related tests** under a parent test function (or a `suite.Suite`)
2. **Use `require` for critical assertions** that should fail fast
3. **Use `assert` for value assertions** that should continue
4. **Mock dependencies** using the Testify mocks in `tests/mocks`
5. **Use external test packages** (`package services_test`, `package handlers_test`, ...) for black-box tests

### Test Structure

```go
func TestUserService(t *testing.T) {
    t.Run("GetProfile - Success", func(t *testing.T) {
        // Arrange
        repo := new(mocks.MockUserRepository)
        mailer := new(mocks.MockMailerService)
        service := services.NewUserService(repo, mailer)
        user := &models.User{ID: 1, Email: "test@example.com"}
        repo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

        // Act
        result, err := service.GetProfile(context.Background(), 1)

        // Assert
        require.NoError(t, err)
        assert.Equal(t, user, result)
        repo.AssertExpectations(t)
    })

    t.Run("GetProfile - Not Found", func(t *testing.T) {
        repo := new(mocks.MockUserRepository)
        service := services.NewUserService(repo, new(mocks.MockMailerService))
        repo.On("GetByID", mock.Anything, uint(999)).Return(&models.User{}, errors.New("not found")).Once()

        result, err := service.GetProfile(context.Background(), 999)

        require.Error(t, err)
        assert.Nil(t, result)
        repo.AssertExpectations(t)
    })
}
```

### Testing Best Practices

1. **One test per feature** - Don't test multiple features in one test
2. **Arrange-Act-Assert (AAA)** - Clear three-part structure
3. **Use table-driven tests** for multiple scenarios
4. **Mock external dependencies** - Database, HTTP, external APIs
5. **Test error cases** - Invalid input, missing data, service errors
6. **Test edge cases** - Boundary conditions, empty inputs
7. **Keep tests independent** - No dependencies between tests
8. **Use meaningful names** - Test names should describe what's being tested

### Table-Driven Tests

```go
func TestValidatePasswordComplexity(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"Valid password", "Secret@1", false},
        {"No special character", "Secret123", true},
        {"Too short", "Se@1", true},
        {"Lowercase only", "lowercaseonly", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... validate tt.password and compare with tt.wantErr
        })
    }
}
```

### Test Coverage Goals

- **Handlers**: 95%+ (critical for API contracts)
- **Services**: 85%+ (core business logic)
- **Repositories**: 90%+ (data access)
- **Middlewares**: 85%+ (security-related)
- **Utils**: 80%+ (utility functions)

The CI pipeline (`.github/workflows/build.yml`) fails the build when total coverage of the tested packages (everything except `cmd`, `docs`, and `tests`) drops below 70%.

---

## Git Workflow

### Branch Naming

```
main                           # Production-ready code
├── feat/feature-name          # New features
├── fix/bug-description        # Bug fixes
├── chore/maintenance-task     # Tooling, dependency and version updates
└── docs/documentation         # Documentation updates
```

Open pull requests against `main`. The CI workflow runs lint, unit tests with coverage, and a 70% coverage gate on every pull request.

### Commit Message Format

Follow conventional commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `style`: Code style changes
- `chore`: Build, dependency updates

**Examples:**

```
feat(auth): implement JWT refresh token
- Add refresh token service
- Update auth handler to support token refresh
- Add tests for refresh token flow

fix(security): hash reset tokens and harden auth flows
- Store only the hash of the reset token
- Add tests for the reset flow

test(auth): broaden auth unit tests with boundary and edge cases
```

### Pull Request Guidelines

1. **Keep PRs focused** - One feature or bug fix per PR
2. **Add clear description** - Explain what and why
3. **Include tests** - All features must have tests
4. **Update documentation** - If API or configuration changes
5. **Request reviews** - Get feedback before merging
6. **Resolve conflicts** - Keep branch up to date with main

`make pre-push` (fmt, vet, lint, test) runs the same checks as CI locally. A gitleaks pre-commit hook is configured in `.pre-commit-config.yaml`.

---

## Documentation

### Code Comments

**Good comments explain WHY, not WHAT:**

```go
// Good - explains the reason
// Validate the access token BEFORE rotating the refresh token. Otherwise an
// attacker holding only a stolen refresh token could burn it and log the
// legitimate user out.
claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)

// Bad - just repeats the code
// Validate the access token
claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)
```

### Function Documentation

```go
// EncryptSecret encrypts plaintext with AES-256-GCM using a key derived from
// secretKey. The result is "enc:v1:" + base64(nonce || ciphertext).
func EncryptSecret(secretKey, plaintext string) (string, error) {
    // implementation
}
```

### Package Documentation

```go
// Command encrypt-setting encrypts a secret setting value (e.g. mail_password)
// with SETTINGS_ENCRYPTION_KEY so it can be stored in the settings table.
package main
```

### API Documentation

The OpenAPI specification is maintained by hand in `docs/swagger.json` (the project does not generate it from code annotations). When you add, change, or remove an endpoint, update `docs/swagger.json` together with the route in `internal/routes/routes.go` and the endpoint list in `README.md`. Swagger UI (`docs/swagger.html`) loads the spec from `/docs/swagger.json` and is only served when `STAGE` is not `prod`.

### Logging Documentation

Follow [docs/logging-standards.md](docs/logging-standards.md) for log fields, event names, and masking of sensitive data.

---

## Performance Guidelines

### Database

1. **Use indexes** on frequently queried columns
2. **Batch operations** when possible
3. **Use transactions** for related operations (see `BeginTx` in the repositories and the refresh-token rotation)
4. **Avoid N+1 queries** - use eager loading
5. **Pagination** for large result sets (`dto.Pagination`, `utils.ParsePageAndLimit`; the default page size is `constants.LIMIT`)

The connection pool is configured in `internal/configs/database.go` and can be tuned with `DB_MAX_OPEN_CONNS` and `DB_MAX_IDLE_CONNS`.

### Caching

The project has no cache layer. If you add one, guard shared state with a mutex (or use a proven library) and use the `ErrCache*` codes in `pkg/apperror` for cache failures.

### Concurrency

1. **Use channels** for goroutine communication
2. **Use sync.Mutex** for shared state (see the in-memory rate limiter)
3. **Avoid goroutine leaks** - always clean up
4. **Use context** for cancellation

---

## Security Guidelines

### Authentication

- Use JWT for stateless authentication: access tokens are valid for 1 hour and carry an `access` scope; the auth middleware only accepts HMAC-signed tokens with that scope
- Refresh tokens are random 60-character strings stored in the database (30-day expiry); they are rotated on every refresh and deleted on logout
- Password reset tokens are valid for 1 hour and only their hash is stored
- Store sensitive data in environment variables (`JWT_KEY`, `SETTINGS_ENCRYPTION_KEY`, database credentials)
- Secret rows of the `settings` table (`mail_password`) are encrypted with AES-256-GCM (`utils.EncryptSecret`)

### Password Security

- Use bcrypt for hashing (never store plaintext)
- Enforce strong passwords with the `password_complexity` validator: at least 8 characters with an uppercase letter, a lowercase letter, a digit, and a special character
- Accounts are locked for 15 minutes after 5 failed logins (`services.MaxFailedAttempts`, `services.LockoutDurationMinutes`)
- The public auth endpoints are rate limited to 10 requests per minute per client IP (`middlewares.RateLimiter`)
- Never reveal whether an email exists (login returns "Invalid credentials"; forgot-password always returns the same message)

### Input Validation

- Validate all user inputs with `binding` tags on DTOs and `utils.TranslateValidationErrors`
- Use parameterized queries (GORM)
- Never log raw sensitive values; use `utils.MaskWithPrefix` and the masking done by the log middleware

### CORS

CORS is handled by `middlewares.CORSMiddleware`, which reads `CORS_ALLOWED_ORIGINS` (comma-separated exact origins, default `http://localhost:5173`) and only echoes the `Origin` header back when it is on the list. Never set it to `*` in production because credentials are allowed.

### Proxies

`TRUSTED_PROXIES` is empty by default, so Gin ignores `X-Forwarded-For` and the client IP cannot be spoofed. Set it to the CIDRs of your reverse proxy or load balancer only when the app runs behind one.

---

## Environment Variables

Copy `.env.example` to `.env` and adjust it. The minimum configuration is:

```bash
# Database
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=password
DB_DATABASE=golang_cms

# JWT (must be at least 32 characters)
JWT_KEY=your-32-character-secret-key-here

# Settings encryption (must be at least 32 characters)
SETTINGS_ENCRYPTION_KEY=your-32-character-encryption-key-here

# Server
PORT=3000
GIN_MODE=debug
RUN_MIGRATE=true
STAGE=local
```

`DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`, `JWT_KEY`, and `SETTINGS_ENCRYPTION_KEY` are required; the server exits at startup if one is missing. Optional variables: `APP_SERVICE`, `APP_VERSION`, `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`. See the [README](README.md#environment-variables) for the full list with defaults.

Mail (`mail_*`) and frontend (`frontend_url`) settings are not environment variables; they live in the `settings` table (key/value) and are seeded by migrations. `mail_password` is stored encrypted with `SETTINGS_ENCRYPTION_KEY` (generate the value with `make encrypt-setting`).

---

## Deployment

### Docker Best Practices

1. Use multi-stage builds for smaller images
2. Run as non-root user
3. Set resource limits
4. Use health checks

The repository `Dockerfile` follows these practices: it builds a static binary in a `golang:1.27-alpine` stage, then copies the binary and the `internal/` directory (which contains the migrations) into an `alpine:3.21` image that runs as a non-root `appuser` and exposes port 3000. The `docs/` directory is not copied, so Swagger is not available in the container unless you add it. The image has no `HEALTHCHECK`; to add one, probe `GET /healthz`.

`docker-compose.yml` only provides local dependencies (MySQL, phpMyAdmin, Mailpit), not the application itself.

### Database Migrations

Migrations are SQL files in `internal/database/migrations`, applied with golang-migrate:

```bash
# Apply migrations on server start (run from the repository root)
RUN_MIGRATE=true go run cmd/server/main.go

# Seed sample users (after the tables exist)
go run cmd/seeder/seeder.go
```

---

## Tools and Dependencies

### Essential Tools

- **Testing**: testify (assert, require, mock, suite); mocks in `tests/mocks` are written by hand
- **HTTP**: gin-gonic/gin
- **Database**: GORM with MySQL (production) and SQLite (tests)
- **JWT**: golang-jwt
- **Validation**: go-playground/validator
- **Logging**: sirupsen/logrus
- **Email**: wneessen/go-mail
- **Config**: joho/godotenv
- **Database Migration**: golang-migrate

### Development Tools

- **Linting**: golangci-lint (configured in `.golangci.yml`)
- **Formatting**: gofmt
- **Testing**: go test, gotestsum
- **Live reload**: Air (`.air.toml`)
- **Secret scanning**: gitleaks (pre-commit hook)

---

## Common Mistakes to Avoid

1. **Not implementing interfaces** - Use interfaces for testability
2. **Ignoring errors** - Always handle and log errors appropriately
3. **No input validation** - Validate at the handler level
4. **Direct database access in handlers** - Use repositories
5. **Mixing concerns** - Keep layers separate
6. **No error types** - Use custom errors for clarity
7. **Inadequate test coverage** - Aim for 80%+ coverage
8. **Hardcoding values** - Use constants and config
9. **No logging** - Log important operations
10. **Poor transaction handling** - Use transactions for multi-step operations

---

## Code Review Checklist

Before submitting a PR:

- [ ] Code follows naming conventions
- [ ] All functions have comments
- [ ] Tests are included and passing
- [ ] Error handling is appropriate
- [ ] No hardcoded values
- [ ] Dependencies are injected
- [ ] Code is DRY (Don't Repeat Yourself)
- [ ] Performance is considered
- [ ] Security is considered
- [ ] Documentation is updated (README, `docs/swagger.json`, this file)
- [ ] Commit messages are clear
- [ ] No unnecessary comments

---

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM Documentation](https://gorm.io/)

---

Last Updated: September 30, 2026

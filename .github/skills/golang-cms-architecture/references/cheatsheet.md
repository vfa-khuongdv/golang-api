# Quick Reference Cheatsheet

## Common Commands

```bash
# Development
make dev              # Start with hot reload (Air)
go run cmd/server/main.go   # Run from the repository root

# Testing
go test ./... -v                  # Run all tests
go test ./... -cover              # With coverage
go test ./... -race               # Race detection
make test-e2e                     # End-to-end tests (tests/e2e)
make test-coverage                # Coverage report (coverage.html)

# Build
make build            # Build binary to ./bin/server

# Code quality
make fmt              # Format code
make vet              # Run go vet
make lint             # Run golangci-lint
make pre-push         # Full check (fmt vet lint test)

# Database
RUN_MIGRATE=true go run cmd/server/main.go  # Apply SQL migrations on startup
go run cmd/seeder/seeder.go  # Seed sample users (tables must exist)
docker-compose up -d mysql   # Start MySQL

# Settings
make encrypt-setting         # Encrypt a secret setting (e.g. mail_password)
```

## Error Handling Patterns

```go
// 400 - Validation (handlers): binding error -> *apperror.ValidationError
utils.RespondWithError(c, utils.TranslateValidationErrors(err, input))

// 400 - Bad request
apperror.NewBadRequestError("Invalid input")

// 404 - Not Found  
apperror.NewNotFoundError("User not found")

// 401 - Unauthorized
apperror.NewUnauthorizedError("Invalid credentials")

// 409 - Conflict
apperror.NewConflictError("Email already exists")

// 429 - Account locked
apperror.NewAccountLockedError("Account is temporarily locked")

// 500 - Internal
apperror.NewInternalServerError("Failed to process")
apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to process", err)
```

## Context Usage

```go
// Handler → Service → Repository
func (h *handler) HandlerFunc(c *gin.Context) {
    ctx := c.Request.Context()  // Already has request_id
    result, err := h.service.Method(ctx, param)
}

// Service
func (s *service) Method(ctx context.Context, param string) (*Model, error) {
    return s.repo.GetByParam(ctx, param)
}

// Repository
func (r *repo) GetByParam(ctx context.Context, param string) (*Model, error) {
    return r.db.WithContext(ctx).First(&model, param)
}
```

## Logging Patterns

```go
// Request-scoped (has request_id)
logger.WithContext(ctx).Infof("Processing %d", id)

// Startup/seeders (no context)
logger.Infof("Server started on %s", port)

// With an event name (adds "event" for filtering)
logger.WithEvent(ctx, logger.EventProfileUpdate).Infof("Updated user %d", id)

// With extra fields
logger.WithContext(ctx).WithField("user_id", id).Info("Updated")
```

## JSON Naming

```go
// ALWAYS use snake_case
type User struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    IsActive  bool      `json:"is_active"`
}
```

## Test Patterns

```go
// AAA Pattern
t.Run("GetProfile - Success", func(t *testing.T) {
    // Arrange
    repo := new(mocks.MockUserRepository)
    repo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{ID: 1}, nil)
    svc := services.NewUserService(repo, new(mocks.MockMailerService))

    // Act
    result, err := svc.GetProfile(context.Background(), 1)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, uint(1), result.ID)
})
```

## Layer Dependencies

```
Handler → Service → Repository → Model
    ↓         ↓         ↓
   DTO      apperror  GORM
```
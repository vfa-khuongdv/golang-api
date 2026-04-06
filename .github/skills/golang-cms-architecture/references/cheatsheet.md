# Quick Reference Cheatsheet

## Common Commands

```bash
# Development
make dev              # Start with hot reload (Air)
go run cmd/server/main.go

# Testing
go test ./... -v                  # Run all tests
go test ./... -cover              # With coverage
go test ./... -race               # Race detection

# Build
make build            # Build binary to ./bin/server

# Code quality
make fmt              # Format code
make vet              # Run go vet
make lint             # Run golangci-lint
make pre-push         # Full check (fmt vet lint test)

# Database
go run cmd/seeder/seeder.go  # Seed data
docker-compose up -d mysql   # Start MySQL
```

## Error Handling Patterns

```go
// 400 - Validation
apperror.NewValidationError("Field is required")

// 404 - Not Found  
apperror.NewNotFoundError("User not found")

// 401 - Unauthorized
apperror.NewUnauthorizedError("Invalid credentials")

// 409 - Conflict
apperror.NewConflictError("Email already exists")

// 500 - Internal
apperror.NewInternalServerError("Failed to process: %w", err)
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
t.Run("Name - Success", func(t *testing.T) {
    // Arrange
    mock := new(mocks.MockDependency)
    svc := NewService(mock)
    
    // Act
    result, err := svc.Method(ctx, input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
})
```

## Layer Dependencies

```
Handler → Service → Repository → Model
    ↓         ↓         ↓
   DTO      apperror  GORM
```
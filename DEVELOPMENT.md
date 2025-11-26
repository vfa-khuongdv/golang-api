# Development Guidelines

This document outlines the coding standards, project structure, and best practices for the Golang CMS project.

## Table of Contents

1. [Project Architecture](#project-architecture)
2. [Code Organization](#code-organization)
3. [Naming Conventions](#naming-conventions)
4. [Interface Design](#interface-design)
5. [Error Handling](#error-handling)
6. [Testing Guidelines](#testing-guidelines)
7. [Git Workflow](#git-workflow)
8. [Documentation](#documentation)

---

## Project Architecture

### Clean Architecture Pattern

This project follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP/REST Layer                       │
│                    (handlers, routes)                    │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    Service Layer                         │
│              (business logic, validation)               │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                  Repository Layer                        │
│            (database access, queries)                   │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                   Data Layer                             │
│                    (Database)                            │
└─────────────────────────────────────────────────────────┘
```

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
│   ├── server/                   # Main application entry point
│   │   └── main.go
│   └── seeder/                   # Database seeder
│       └── seeder.go
├── internal/                     # Private application code
│   ├── configs/                  # Configuration management
│   ├── constants/                # Application constants
│   ├── database/                 # Database setup
│   │   ├── migrations/           # Migration files
│   │   └── seeders/              # Seeder implementations
│   ├── dto/                      # Data transfer objects for request and response
│   ├── handlers/                 # HTTP handlers/controllers
│   ├── middlewares/              # HTTP middlewares
│   ├── models/                   # Data models
│   ├── repositories/             # Data access layer
│   ├── routes/                   # Route definitions
│   ├── services/                 # Business logic layer
│   └── utils/                    # Utility functions
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

### Responsibilities by Layer

#### **Models (internal/models/)**
- Define data structures
- Represent database tables
- Include validation tags
- Keep models simple - no business logic

#### **Repositories (internal/repositories/)**
- Handle all database operations
- Implement CRUD operations
- Return domain entities
- No business logic - only data access
- Implement interfaces for mockability

#### **Services (internal/services/)**
- Contain business logic
- Validate data
- Orchestrate between repositories
- Handle errors appropriately
- Return clean data to handlers

#### **Handlers (internal/handlers/)**
- Parse HTTP requests
- Call appropriate services
- Return HTTP responses
- Handle response formatting
- Minimal business logic

#### **Middlewares (internal/middlewares/)**
- Authentication/authorization
- Request logging
- CORS handling
- Error handling
- Input validation

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
  func GetUser()           // Exported
  func getUserByEmail()    // Unexported
  ```

- **Constants**: All uppercase with underscores
  ```go
  const (
      MAX_RETRY_COUNT = 3
      DEFAULT_TIMEOUT = 30
  )
  ```

- **Interfaces**: End with `I` prefix or descriptive name
  ```go
  type IUserService interface {}    // Prefix pattern
  type UserService interface {}      // Descriptive pattern
  ```

- **Variables**: CamelCase for exported, camelCase for unexported
  ```go
  var GlobalConfig Config           // Exported
  var userRepository UserRepository // Unexported
  ```

### Naming Patterns by Layer

#### **Models**
- Use singular names for structs
- Use descriptive field names
- Use JSON tags for serialization in **snake_case** format for API responses

```go
type User struct {
    ID        uint      `json:\"id\"`
    Email     string    `json:\"email\"`
    Name      string    `json:\"name\"`
    CreatedAt time.Time `json:\"created_at\"`
    UpdatedAt time.Time `json:\"updated_at\"`
}
```

**Note:** All JSON field names in API responses must use snake_case (e.g., `created_at`, `user_id`, `is_active`) for consistency with REST API conventions.

#### **Interfaces**
- Describe the contract clearly
- Use `I` prefix for repository/service interfaces
- Keep interfaces small and focused

```go
type IUserRepository interface {
    GetByID(id uint) (*User, error)
    GetByEmail(email string) (*User, error)
    Create(user *User) error
}
```

#### **Methods**
- Use verb-based names for actions
- Use `Get` for retrieval
- Use `Create`, `Update`, `Delete` for modifications
- Use `Is`, `Has` for boolean checks

```go
func (s *UserService) GetUser(id uint) (*User, error)
func (s *UserService) CreateUser(data *CreateUserRequest) error
func (s *UserService) IsValidEmail(email string) bool
```

#### **Test Functions**
- Use format: `TestFunctionName` or `TestFunctionName/subtest`
- Group related tests under parent test function
- Use descriptive subtest names

```go
func TestUserService(t *testing.T) {
    t.Run("CreateUser - Success", func(t *testing.T) {
        // test code
    })
    t.Run("CreateUser - Validation Error", func(t *testing.T) {
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
   type UserService struct {
       repo IUserRepository
   }

   // Bad - depends on concrete type
   type UserService struct {
       repo *UserRepository
   }
   ```

2. **Single Responsibility**
   ```go
   // Good - focused interface
   type IUserRepository interface {
       GetByID(id uint) (*User, error)
       Create(user *User) error
   }

   // Bad - too many responsibilities
   type IDatabase interface {
       Query(sql string) Results
       Execute(sql string) error
       GetUser(id uint) (*User, error)
       CreateUser(user *User) error
   }
   ```

3. **Interface Segregation**
   ```go
   // Good - small, specific interfaces
   type Reader interface {
       Read() ([]byte, error)
   }
   type Writer interface {
       Write([]byte) error
   }

   // Bad - large, monolithic interface
   type FileHandler interface {
       Read() ([]byte, error)
       Write([]byte) error
       Delete() error
       // ... many more methods
   }
   ```

### Handler Interface Pattern

```go
// Define interface
type IUserHandler interface {
    GetUser(c *gin.Context)
    CreateUser(c *gin.Context)
    UpdateUser(c *gin.Context)
    DeleteUser(c *gin.Context)
}

// Implement with dependency injection
type UserHandler struct {
    userService IUserService
    jwtService  IJWTService
}

func NewUserHandler(userService IUserService, jwtService IJWTService) IUserHandler {
    return &UserHandler{
        userService: userService,
        jwtService:  jwtService,
    }
}
```

---

## Error Handling

### Custom Error Structure

Use `apperror` package for consistent error handling:

```go
type AppError struct {
    HttpStatusCode int    `json:"http_status_code"`
    Code           int    `json:"code"`
    Message        string `json:"message"`
    UnderlyingErr  error  `json:"-"`
}
```

### Error Creation

```go
// Create new error
err := apperror.NewInternalError("Database error")

// Wrap existing error
err := apperror.Wrap(
    http.StatusInternalServerError,
    apperror.ErrInternal,
    "Failed to create user",
    originalError,
)

// Specific error types
apperror.NewBadRequestError("Invalid input")
apperror.NewUnauthorizedError("Invalid credentials")
apperror.NewForbiddenError("Access denied")
apperror.NewNotFoundError("User not found")
apperror.NewInvalidPasswordError("Wrong password")
apperror.NewValidationError("Validation failed")
apperror.NewInternalError("Server error")
```

### Error Handling in Handlers

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    userID, err := strconv.ParseUint(id, 10, 32)
    if err != nil {
        // Validation error
        utils.RespondWithError(c, apperror.NewBadRequestError("Invalid user ID"))
        return
    }

    user, err := h.userService.GetUser(uint(userID))
    if err != nil {
        // Service returns app errors
        utils.RespondWithError(c, err)
        return
    }

    utils.RespondWithOK(c, http.StatusOK, gin.H{
        "data": user,
    })
}
```

### Error Handling in Services

```go
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // Validation
    if req.Email == "" {
        return nil, apperror.NewValidationError("Email is required")
    }

    // Check duplicate
    existing, err := s.repo.FindByEmail(req.Email)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, apperror.Wrap(
            http.StatusInternalServerError,
            apperror.ErrInternal,
            "Failed to check existing user",
            err,
        )
    }
    if existing != nil {
        return nil, apperror.NewBadRequestError("Email already exists")
    }

    // Create user
    user := &User{Email: req.Email}
    if err := s.repo.Create(user); err != nil {
        return nil, apperror.Wrap(
            http.StatusInternalServerError,
            apperror.ErrInternal,
            "Failed to create user",
            err,
        )
    }

    return user, nil
}
```

---

## Testing Guidelines

### Test Organization

All tests should follow the **Testify** framework with proper grouping:

1. **Group related tests** under a parent test function
2. **Use `require` for critical assertions** that should fail fast
3. **Use `assert` for value assertions** that should continue
4. **Mock dependencies** using Testify mock package

### Test Structure

```go
func TestUserService(t *testing.T) {
    // Setup
    mockRepo := new(mocks.MockUserRepository)
    service := services.NewUserService(mockRepo)

    // Group related test cases
    t.Run("CreateUser - Success", func(t *testing.T) {
        // Arrange
        mockRepo.On("FindByEmail", "test@example.com").Return(nil, sql.ErrNoRows)
        mockRepo.On("Create", mock.Anything).Return(nil)

        // Act
        user, err := service.CreateUser(&CreateUserRequest{
            Email: "test@example.com",
        })

        // Assert
        require.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "test@example.com", user.Email)
        mockRepo.AssertExpectations(t)
    })

    t.Run("CreateUser - Duplicate Email", func(t *testing.T) {
        // Arrange
        existing := &User{ID: 1, Email: "test@example.com"}
        mockRepo.On("FindByEmail", "test@example.com").Return(existing, nil)

        // Act
        user, err := service.CreateUser(&CreateUserRequest{
            Email: "test@example.com",
        })

        // Assert
        require.Error(t, err)
        assert.Nil(t, user)
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
func TestValidateEmail(t *testing.T) {
    t.Run("Email Validation", func(t *testing.T) {
        tests := []struct {
            name      string
            email     string
            wantValid bool
        }{
            {"Valid email", "user@example.com", true},
            {"Missing @", "userexample.com", false},
            {"Missing domain", "user@", false},
            {"Empty string", "", false},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                valid := ValidateEmail(tt.email)
                assert.Equal(t, tt.wantValid, valid)
            })
        }
    })
}
```

### Test Coverage Goals

- **Handlers**: 95%+ (critical for API contracts)
- **Services**: 85%+ (core business logic)
- **Repositories**: 90%+ (data access)
- **Utils**: 80%+ (utility functions)

---

## Git Workflow

### Branch Naming

```
main                           # Production-ready code
├── feature/feature-name       # New features
├── bugfix/bug-description     # Bug fixes
├── hotfix/urgent-issue        # Production hotfixes
└── docs/documentation         # Documentation updates
```

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

fix(user): correct email validation regex
- Fix edge case in email validation
- Add test for edge case

test(mfa): add comprehensive MFA tests
- Add tests for setup and verification
- Add tests for edge cases
```

### Pull Request Guidelines

1. **Keep PRs focused** - One feature or bug fix per PR
2. **Add clear description** - Explain what and why
3. **Include tests** - All features must have tests
4. **Update documentation** - If API or configuration changes
5. **Request reviews** - Get feedback before merging
6. **Resolve conflicts** - Keep branch up to date with main

---

## Documentation

### Code Comments

**Good comments explain WHY, not WHAT:**

```go
// Good - explains the reason
// We use bcrypt cost 12 for security-performance balance
// Lower costs are vulnerable, higher costs cause user-visible delays
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(password),
    12,
)

// Bad - just repeats the code
// Hash the password with cost 12
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(password),
    12,
)
```

### Function Documentation

```go
// CreateUser creates a new user in the system
//
// It validates the request, checks for duplicate emails,
// and hashes the password before storing.
//
// Returns ErrDuplicate if email already exists.
// Returns ErrValidation if request is invalid.
// Returns ErrInternal on database errors.
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // implementation
}
```

### Package Documentation

```go
// Package services provides business logic for the application.
//
// It contains the following services:
// - UserService: User management
// - AuthService: Authentication and authorization
// - MFAService: Multi-factor authentication
package services
```

### README in Each Package

Add `README.md` to complex packages:

```markdown
# Services Package

This package contains all business logic for the application.

## Services

### UserService
Handles user management operations like creation, retrieval, and updates.

### AuthService  
Handles authentication, login, and token management.
```

### API Documentation

Maintain OpenAPI/Swagger documentation for all endpoints:

```go
// @Summary Get user by ID
// @Description Get detailed user information
// @Tags users
// @Security Bearer
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} AppError "User not found"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    // implementation
}
```

---

## Performance Guidelines

### Database

1. **Use indexes** on frequently queried columns
2. **Batch operations** when possible
3. **Use transactions** for related operations
4. **Avoid N+1 queries** - use eager loading
5. **Pagination** for large result sets

### Caching

```go
// Cache frequently accessed data
type UserCache struct {
    users map[uint]*User
    mu    sync.RWMutex
}

func (c *UserCache) Get(id uint) (*User, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    user, exists := c.users[id]
    return user, exists
}
```

### Concurrency

1. **Use channels** for goroutine communication
2. **Use sync.Mutex** for shared state
3. **Avoid goroutine leaks** - always clean up
4. **Use context** for cancellation

---

## Security Guidelines

### Authentication

- Use JWT for stateless authentication
- Include expiration times in tokens
- Rotate refresh tokens regularly
- Store sensitive data in environment variables

### Password Security

- Use bcrypt for hashing (never store plaintext)
- Enforce strong password requirements
- Implement rate limiting on login attempts
- Consider 2FA for sensitive operations

### Input Validation

- Validate all user inputs
- Sanitize before storing in database
- Use parameterized queries
- Validate file uploads

### CORS

```go
config := cors.Config{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

---

## Environment Variables

Create `.env.example` file for required variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=golang_cms

# JWT
JWT_SECRET_KEY=your_secret_key_here
JWT_EXPIRATION_HOURS=24
JWT_REFRESH_EXPIRATION_HOURS=168

# Server
SERVER_PORT=8080
SERVER_ENV=development

# Email
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your_email@gmail.com
MAIL_PASSWORD=app_password
MAIL_FROM=noreply@example.com

# Frontend
FRONTEND_URL=http://localhost:3000
```

---

## Deployment

### Docker Best Practices

1. Use multi-stage builds for smaller images
2. Run as non-root user
3. Set resource limits
4. Use health checks

```dockerfile
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /app/app .
USER appuser
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["./app"]
```

### Database Migrations

Always use migration tools:

```bash
# Run migrations
go run cmd/server/main.go migrate

# Rollback migrations  
go run cmd/server/main.go migrate-down
```

---

## Tools and Dependencies

### Essential Tools

- **Testing**: testify, mockery
- **HTTP**: gin-gonic/gin
- **Database**: GORM
- **JWT**: golang-jwt
- **Validation**: go-playground/validator
- **Logging**: sirupsen/logrus
- **Config**: joho/godotenv
- **Database Migration**: golang-migrate

### Development Tools

- **Linting**: golangci-lint
- **Formatting**: gofmt
- **Testing**: go test
- **Benchmarking**: go test -bench

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
- [ ] Documentation is updated
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

Last Updated: November 21, 2025

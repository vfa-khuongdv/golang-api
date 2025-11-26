# Golang CMS - Copilot Instructions

## Repository Overview

This is a Go-based CMS REST API with user authentication, MFA (TOTP), JWT tokens, and refresh tokens. The project implements clean architecture with handlers → services → repositories → models layers. All code uses testify for testing and the apperror package for standardized error handling.

**Key Technologies:** Go 1.21+, Gin framework, GORM, MySQL, JWT, Testify, Docker

## Project Layout & Architecture

```
golang-cms/
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
├── pkg/                        # Reusable packages
│   ├── apperror/               # Error codes and error types
│   ├── logger/                 # Logging utility
│   ├── mailer/                 # Email/SMTP service
│   └── migrator/               # Database migration runner
├── tests/mocks/                # Mock implementations for testing
├── DEVELOPMENT.md              # Detailed development guidelines
├── TESTING.md                  # Testing standards and patterns
├── Dockerfile                  # Docker container configuration
├── docker-compose.yml          # Multi-container orchestration
├── go.mod                      # Go module definition
├── Makefile                    # Build and run commands
└── README.md                   # Project documentation
```

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

## Key Architectural Patterns

### Layer Responsibilities
- **Handlers** (internal/handlers/): Parse HTTP requests, call services, return HTTP responses, minimal business logic
- **Services** (internal/services/): Business logic, validation, orchestrate repositories, return errors via apperror
- **Repositories** (internal/repositories/): Database operations, return domain entities, implement interfaces
- **Models** (internal/models/): Domain objects, GORM tags, JSON serialization tags (snake_case), validation
- **Middlewares** (internal/middlewares/): Auth/authorization, request logging, CORS, error handling

### Dependencies & Interfaces
- **All layers depend on abstractions (interfaces), never concrete implementations**
- Services receive repositories via constructor (dependency injection)
- Handlers receive services via constructor
- Interfaces use `I` prefix: `IUserService`, `IUserRepository`
- Keep interfaces small and focused (Interface Segregation Principle)
- Test mocks use testify/mock package

**Good Dependency Injection Pattern:**
```go
type UserService struct {
    repo IUserRepository  // Depends on interface, not concrete type
}

func NewUserService(repo IUserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### Naming Conventions (see DEVELOPMENT.md for full details)

**Packages & Files:**
- Use lowercase, single-word names: `handlers`, `services`, `repositories`
- Test files: `user_service_test.go`, `*_integration_test.go`

**Functions/Methods:**
- Use verb-based names for actions
- Use `Get` for retrieval, `Create`/`Update`/`Delete` for modifications, `Is`/`Has` for booleans
- Example: `GetUser(id)`, `CreateUser(req)`, `IsValidEmail(email)`

**Constants:**
- All uppercase with underscores: `MAX_RETRY_COUNT`, `DEFAULT_TIMEOUT`

**Test Functions:**
- Format: `TestFunctionName` or grouped under parent: `TestUserService` with subtests
- Use descriptive subtest names: `t.Run("CreateUser - Success", ...)`

**JSON Tags:**
- Use snake_case for API responses (REST conventions): `created_at`, `user_id`, `is_active`
- Example: 
  ```go
  type User struct {
      ID        uint      `json:"id"`
      Email     string    `json:"email"`
      CreatedAt time.Time `json:"created_at"`
  }
  ```

### Error Handling
- **All errors use** `github.com/vfa-khuongdv/golang-cms/pkg/apperror` package
- **Validation errors**: `apperror.NewValidationError()` (HTTP 400)
- **Not found errors**: `apperror.NewNotFoundError()` (HTTP 404)
- **Auth errors**: `apperror.NewUnauthorizedError()` (HTTP 401)
- **Conflict errors**: `apperror.NewConflictError()` (HTTP 409)
- **Server errors**: Wrap with context `apperror.Wrap(statusCode, code, message, originalErr)`
- **HTTP status mapping**: 400 (validation), 401 (auth), 403 (forbidden), 404 (not found), 409 (conflict), 500 (server error)

### Authentication
- JWT tokens generated by JWTService
- Tokens expire (configurable, typically 15 minutes)
- Refresh tokens stored in database via RefreshTokenRepository
- MFA (TOTP) implemented with TotpService and MfaRepository
- AuthMiddleware validates JWT on protected routes

## Testing Standards

See **TESTING.md** for comprehensive testing guidelines. Key requirements:

### Test Pyramid
- **Unit Tests (80%)**: Fast, isolated, single function/method
- **Integration Tests (15%)**: Multiple components, external services mocked
- **E2E Tests (5%)**: Full user flows, real database

### Testify Framework Usage
**Use three main packages:**
1. **assert** - Soft assertions (continue on failure): `assert.Equal(t, expected, actual)`
2. **require** - Hard assertions (stop on failure): `require.NoError(t, err)`
3. **mock** - Object mocking: `mockRepo.On("Method", args).Return(result, nil)`

### All Tests Must
- Use testify/assert and testify/require packages (never t.Errorf)
- Follow **AAA pattern**: Arrange (setup), Act (execute), Assert (verify)
- Group related tests using `t.Run()` with descriptive names
- Use `require` for critical checks, `assert` for value checks
- Mock external dependencies (repositories, services)
- Include both success and failure test cases
- Never use hardcoded values (use named constants)
- Call `mockRepo.AssertExpectations(t)` after each mock usage

### Test File Structure
```go
func TestUserService(t *testing.T) {
    // Setup phase - creates all necessary objects
    mockRepo := new(mocks.MockUserRepository)
    service := services.NewUserService(mockRepo)

    t.Run("CreateUser - Success", func(t *testing.T) {
        // Arrange - set up test data and expectations
        user := &User{Email: "test@example.com", Name: "Test User"}
        mockRepo.On("Create", mock.MatchedBy(func(u *User) bool {
            return u.Email == user.Email
        })).Return(nil).Once()

        // Act - execute the function being tested
        result, err := service.CreateUser(user)

        // Assert - verify the results
        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, user.Email, result.Email)
        mockRepo.AssertExpectations(t)
    })
    
    t.Run("CreateUser - Validation Error", func(t *testing.T) {
        // Arrange
        user := &User{Email: "", Name: "Test"}
        
        // Act
        result, err := service.CreateUser(user)
        
        // Assert
        assert.Error(t, err)
        assert.Nil(t, result)
    })
}
```

### Repository Testing
- Use in-memory SQLite (`sqlite.Open(":memory:")`) for unit tests
- Create test fixtures before each test
- Clean up database state after each test (use `TearDownTest()`)
- Never use real MySQL database in unit tests
- Use testify suite pattern for shared setup/teardown

### Handler Testing (Gin)
- Set `gin.SetMode(gin.TestMode)` at start
- Mock service layer dependencies
- Use `httptest.NewRecorder()` and `gin.CreateTestContext()`
- Verify both HTTP status code and response body
- Test both success and error paths

### Service Testing
- Mock repository layer dependencies
- Test business logic in isolation
- Verify error handling with different error types
- Test validation logic before repository calls

### Integration Testing
- Use real database (in-memory SQLite)
- Setup multiple services working together
- Test complete workflows (e.g., auth flow, user creation)
- Place in `*_integration_test.go` files

### Coverage Goals by Layer
- **Handlers**: 95%+ coverage required
- **Services**: 85%+ coverage required
- **Repositories**: 90%+ coverage required
- **Middlewares**: 85%+ coverage required
- **Utils**: 80%+ coverage required
- Run: `go test ./... --cover` to verify

## Important Implementation Rules

### Always DO (from DEVELOPMENT.md & TESTING.md)
- Use dependency injection for testability
- Handle all errors explicitly (never `_ = err`)
- Use apperror package for all application errors
- Write tests immediately after writing code
- Verify all tests pass before committing: `go test ./... -v --cover`
- Use meaningful variable and function names
- Add comments explaining "why", not "what" (code explains what)
- Group related test functions with `t.Run()` subtests
- Use `require` for critical assertions, `assert` for value checks
- Mock all external dependencies in unit tests
- Update DEVELOPMENT.md and TESTING.md when adding new patterns
- Follow AAA pattern in all tests: Arrange, Act, Assert
- Use table-driven tests for multiple scenarios
- Keep interfaces small and focused (1-3 methods)
- Clean up resources in test teardowns

### Never DO
- Ignore errors silently
- Use global variables
- Mix concerns between layers (business logic in handlers)
- Hardcode configuration values (use environment variables)
- Create test files without grouped subtests (use `t.Run()`)
- Use internal packages outside internal/
- Bypass AuthMiddleware on protected endpoints
- Store passwords in plain text
- Log sensitive information
- Test multiple things in one test
- Write tests that depend on each other
- Skip error handling tests
- Use time-based assertions (flaky tests)
- Mock internal dependencies (only external ones)
- Have complex setup in tests
- Mix unit and integration tests in same function

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

Follow this step-by-step workflow when implementing new features:

1. **Start with the model:** Define domain model with validation tags (GORM and JSON tags)
2. **Add repository layer:** Define interface first, implement GORM operations, write unit tests with in-memory SQLite
3. **Add service layer:** Implement business logic, validate inputs, orchestrate repositories, write tests with mocked repositories
4. **Add handler layer:** Accept requests, validate input format, call service, return responses in correct HTTP format, write handler tests
5. **Add routes:** Register handler in internal/routes/routes.go
6. **Run all tests:** `go test ./... -v --cover` must pass and meet coverage targets
7. **Update documentation:** If introducing new patterns, update DEVELOPMENT.md and TESTING.md

### Example: Adding a New Feature

```go
// 1. Define Model (internal/models/post.go)
type Post struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Title     string    `json:"title" validate:"required,min=3"`
    Content   string    `json:"content" validate:"required"`
    UserID    uint      `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 2. Repository Interface (internal/repositories/post_repository.go)
type IPostRepository interface {
    Create(post *Post) error
    GetByID(id uint) (*Post, error)
    GetAll() ([]*Post, error)
    Update(post *Post) error
    Delete(id uint) error
}

// 3. Service (internal/services/post_service.go)
type PostService struct {
    repo IPostRepository
}

func (s *PostService) CreatePost(req *CreatePostRequest) (*Post, error) {
    // Validate
    if req.Title == "" {
        return nil, apperror.NewValidationError("Title is required")
    }
    
    // Create
    post := &Post{Title: req.Title, Content: req.Content, UserID: req.UserID}
    if err := s.repo.Create(post); err != nil {
        return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternal, "Failed to create post", err)
    }
    return post, nil
}

// 4. Handler (internal/handlers/post_handler.go)
type PostHandler struct {
    postService IPostService
}

func (h *PostHandler) CreatePost(c *gin.Context) {
    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.RespondWithError(c, apperror.NewBadRequestError(err.Error()))
        return
    }
    
    post, err := h.postService.CreatePost(&req)
    if err != nil {
        utils.RespondWithError(c, err)
        return
    }
    
    utils.RespondWithOK(c, http.StatusCreated, post)
}

// 5. Tests (internal/handlers/post_handler_test.go)
func TestPostHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    t.Run("CreatePost - Success", func(t *testing.T) {
        // Arrange
        mockService := new(mocks.MockPostService)
        post := &Post{ID: 1, Title: "Test", UserID: 1}
        mockService.On("CreatePost", mock.Anything).Return(post, nil)
        
        handler := handlers.NewPostHandler(mockService)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        
        body, _ := json.Marshal(map[string]any{"title": "Test", "content": "Test content"})
        c.Request, _ = http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(body))
        c.Request.Header.Set("Content-Type", "application/json")
        
        // Act
        handler.CreatePost(c)
        
        // Assert
        require.Equal(t, http.StatusCreated, w.Code)
        mockService.AssertExpectations(t)
    })
}
```

## Debugging Tips

- Enable verbose logging by checking logger.go implementation
- Use `go test -run TestName -v` to run specific tests
- Use `go test -race` to detect race conditions
- Check `internal/configs/database_test.go` for test database setup patterns
- Check `tests/mocks/` for mock implementation examples
- View swagger documentation at `/docs` endpoint when server is running

Trust these instructions. Only search the codebase if you find inconsistencies between these instructions and actual implementation.

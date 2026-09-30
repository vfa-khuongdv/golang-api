# Testing Standards

Comprehensive testing guidelines for the Golang CMS project using Testify framework.

## Table of Contents

1. [Testing Architecture](#testing-architecture)
2. [Testify Framework](#testify-framework)
3. [Unit Testing](#unit-testing)
4. [Integration Testing](#integration-testing)
5. [Mocking Strategy](#mocking-strategy)
6. [Test Organization](#test-organization)
7. [Coverage Goals](#coverage-goals)
8. [Common Test Patterns](#common-test-patterns)
9. [Performance Testing](#performance-testing)
10. [Best Practices Summary](#best-practices-summary)

---

## Testing Architecture

### Test Pyramid

```
       🧪 E2E Tests (5%)
      /              \
     /  Integration   \  (15%)
    /    Tests         \
   /___________________ \
  /                      \     
 /    Unit Tests (80%)    \
/__________________________\
```

- **Unit Tests** (80%): Fast, isolated, single function/method, colocated with the code
- **Integration Tests** (15%): Several real components together (for example repositories against in-memory SQLite)
- **E2E Tests** (5%): The full router and all layers in `tests/e2e`, using in-memory SQLite instead of MySQL

---

## Testify Framework

### Core Packages

We use three main Testify packages:

#### 1. **assert** - Soft Assertions

```go
import "github.com/stretchr/testify/assert"

assert.Equal(t, 5, result)           // Continue on failure
assert.True(t, condition)            // Continue on failure
assert.NotNil(t, value)              // Continue on failure
```

**When to use:** Value checks that don't stop test execution

#### 2. **require** - Hard Assertions

```go
import "github.com/stretchr/testify/require"

require.NoError(t, err)              // Stop on failure
require.NotNil(t, value)             // Stop on failure
require.True(t, condition)           // Stop on failure
```

**When to use:** Critical checks that must pass to continue

#### 3. **mock** - Object Mocking

```go
import "github.com/stretchr/testify/mock"

mockRepo := new(mocks.MockUserRepository)
mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{}, nil)
defer mockRepo.AssertExpectations(t)
```

**When to use:** Simulating dependencies and external calls

### Common Assertions

```go
// Equality
assert.Equal(t, expected, actual)           // Deep equality
assert.NotEqual(t, unexpected, actual)      // Not equal
assert.EqualValues(t, expected, actual)     // Convert types

// Nil checks
assert.Nil(t, value)                        // Should be nil
assert.NotNil(t, value)                     // Should not be nil

// Boolean
assert.True(t, value)                       // Should be true
assert.False(t, value)                      // Should be false

// Collections
assert.Len(t, collection, 5)                // Check length
assert.Empty(t, collection)                 // Check empty
assert.NotEmpty(t, collection)              // Check not empty
assert.Contains(t, collection, element)     // Element in collection

// String
assert.Contains(t, "hello world", "world")  // Substring exists
assert.NotContains(t, "hello", "world")     // Substring missing

// Errors
assert.Error(t, err)                        // Should have error
assert.NoError(t, err)                      // Should have no error
assert.ErrorContains(t, err, "message")     // Error contains message

// Type
assert.IsType(t, (*models.User)(nil), result)      // Check type
assert.Implements(t, (*Reader)(nil), obj)   // Implements interface
```

---

## Unit Testing

### Test Structure (AAA Pattern)

```go
package services_test

func TestUserService(t *testing.T) {
    t.Run("GetProfile - Success", func(t *testing.T) {
        // Arrange - set up mocks, service, and expectations
        repo := new(mocks.MockUserRepository)
        mailer := new(mocks.MockMailerService)
        service := services.NewUserService(repo, mailer)
        user := &models.User{ID: 1, Email: "test@example.com", Name: "Test User"}
        repo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

        // Act - execute the function being tested
        result, err := service.GetProfile(context.Background(), 1)

        // Assert - verify the results
        require.NoError(t, err)
        assert.Equal(t, user, result)
        repo.AssertExpectations(t)
    })
}
```

Service tests can also use `suite.Suite` (see `internal/services/user_service_test.go`): create the mocks and service in `SetupTest` and call `AssertExpectations` in `TearDownTest`.

### Handler Testing

Handlers are concrete structs (`handlers.NewUserHandler(userService, mailerService)`). Tests build a Gin test context, set the authenticated user ID the way `AuthMiddleware` does (`c.Set("UserID", uint(1))`), and call the handler method directly. Call `utils.InitValidator()` first so the custom binding tags (`password_complexity`, `not_blank`, `valid_birthday`) are registered.

```go
package handlers_test

func TestGetProfile(t *testing.T) {
    gin.SetMode(gin.TestMode)
    utils.InitValidator()

    t.Run("GetProfile - Success", func(t *testing.T) {
        // Arrange
        userService := new(mocks.MockUserService)
        mailerService := new(mocks.MockMailerService)
        handler := handlers.NewUserHandler(userService, mailerService)

        user := &models.User{ID: 1, Email: "test@example.com", Name: "Test User"}
        userService.On("GetProfile", mock.Anything, uint(1)).Return(user, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
        c.Set("UserID", uint(1))

        // Act
        handler.GetProfile(c)

        // Assert
        require.Equal(t, http.StatusOK, w.Code)
        var response map[string]any
        require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
        assert.Equal(t, "test@example.com", response["email"])
        userService.AssertExpectations(t)
    })

    t.Run("GetProfile - Not Found", func(t *testing.T) {
        userService := new(mocks.MockUserService)
        handler := handlers.NewUserHandler(userService, new(mocks.MockMailerService))
        userService.On("GetProfile", mock.Anything, uint(1)).
            Return((*models.User)(nil), apperror.NewNotFoundError("User not found"))

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
        c.Set("UserID", uint(1))

        handler.GetProfile(c)

        require.Equal(t, http.StatusNotFound, w.Code)
        assert.JSONEq(t, `{"code":1001,"message":"User not found"}`, w.Body.String())
    })
}
```

Responses are written by `utils.RespondWithOK` / `utils.RespondWithError` as plain JSON bodies: `RespondWithOK` serializes the body you pass (there is no `data` envelope), and errors are `{"code": ..., "message": ...}` (plus `fields` for validation errors).

### Repository Testing

Repository tests run against an in-memory SQLite database created with GORM; only the models the test needs are migrated.

```go
package repositories_test

// setupUserTestDB creates an in-memory SQLite database for testing
func setupUserTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    require.NoError(t, db.AutoMigrate(&models.User{}))
    return db
}

func TestUserRepository(t *testing.T) {
    t.Run("Create and GetByID", func(t *testing.T) {
        // Arrange
        db := setupUserTestDB(t)
        repo := repositories.NewUserRepository(db)
        ctx := context.Background()
        user := &models.User{Name: "Test", Email: "test@example.com", Password: "hashed", Gender: 1}

        // Act
        created, err := repo.Create(ctx, user)
        require.NoError(t, err)
        found, err := repo.GetByID(ctx, created.ID)

        // Assert
        require.NoError(t, err)
        assert.Equal(t, "test@example.com", found.Email)
    })

    t.Run("GetByID - Not Found", func(t *testing.T) {
        db := setupUserTestDB(t)
        repo := repositories.NewUserRepository(db)

        result, err := repo.GetByID(context.Background(), 9999)

        assert.Error(t, err)
        assert.Nil(t, result)
    })
}
```

### Service Testing

Services are tested with mocked repositories and collaborating services. `NewAuthService` takes the user repository, the refresh token service, and the JWT service:

```go
package services_test

func TestAuthService(t *testing.T) {
    t.Run("Login - Invalid Credentials", func(t *testing.T) {
        // Arrange
        userRepo := new(mocks.MockUserRepository)
        refreshTokenService := new(mocks.MockRefreshTokenService)
        jwtService := new(mocks.MockJWTService)
        service := services.NewAuthService(userRepo, refreshTokenService, jwtService)

        hashed, _ := utils.HashPassword("correct-password")
        user := &models.User{ID: 1, Email: "test@example.com", Password: hashed}
        userRepo.On("FindByField", mock.Anything, "email", "test@example.com").Return(user, nil)
        userRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

        // Act
        result, err := service.Login(context.Background(), "test@example.com", "wrong-password", "127.0.0.1")

        // Assert
        require.Error(t, err)
        assert.Nil(t, result)
        appErr, ok := apperror.ToAppError(err)
        require.True(t, ok)
        assert.Equal(t, apperror.ErrInvalidPassword, appErr.Code)
        userRepo.AssertExpectations(t)
    })
}
```

---

## Integration Testing

Integration-style tests live in `tests/e2e`. They start the real router (`routes.SetupRouter`) with every real repository and service, backed by an in-memory SQLite database, and send requests through `router.ServeHTTP`. Only the SMTP server is not exercised. Run them with `make test-e2e` or `go test ./tests/e2e/... -v`.

### End-to-End Setup

`tests/e2e/setup_test.go` provides `setupTestRouter()`, which:

- sets `JWT_KEY` and `SETTINGS_ENCRYPTION_KEY` for the test process
- opens a shared in-memory SQLite database (capped at one connection) and migrates `User`, `RefreshToken`, and `Setting`
- calls `utils.InitValidator()` and returns the configured `*gin.Engine` together with the `*gorm.DB`

```go
package e2e

func TestAuthLogin(t *testing.T) {
    router, db := setupTestRouter()

    // Create a user directly in the database
    hashed, _ := utils.HashPassword("password123")
    user := models.User{Name: "Test User", Email: "test_login@example.com", Password: hashed, Gender: 1}
    require.NoError(t, db.Create(&user).Error)

    t.Run("Login - Success", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "email":    "test_login@example.com",
            "password": "password123",
        })

        w := httptest.NewRecorder()
        req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")

        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        var response dto.LoginResponse
        require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
        assert.NotEmpty(t, response.AccessToken.Token)
        assert.NotEmpty(t, response.RefreshToken.Token)
    })
}
```

The current e2e suites cover login, logout, refresh token, forgot/reset password, change password, get/update profile, empty request bodies, rate limiting, the health check, and the version endpoint.

---

## Mocking Strategy

### Shared Mocks

Reusable mocks live in `tests/mocks` and are written by hand on top of `testify/mock`:

| Mock | Mocks |
|------|-------|
| `MockUserRepository` | `repositories.UserRepository` |
| `MockRefreshTokenRepository` | `repositories.RefreshTokenRepository` |
| `MockSettingRepository` | `repositories.SettingRepository` |
| `MockAuthService` | `services.AuthService` |
| `MockUserService` | `services.UserService` |
| `MockRefreshTokenService` | `services.RefreshTokenService` |
| `MockJWTService` | `services.JWTService` |
| `MockMailerService` | `services.MailerService` |
| `MockDB` / `MockTx` | GORM transaction helpers |

When you add a repository or service interface, add a matching mock next to these.

```go
// A mock is a struct embedding mock.Mock; each method records the call
type MockMailerService struct {
    mock.Mock
}

func (m *MockMailerService) SendMailForgotPassword(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

### Mocking with Arguments

```go
// Every service and repository method receives a context first
repo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{ID: 1}, nil)

// Match argument by type
repo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

// Match argument by function
repo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
    return u.Email != "" && len(u.Name) > 0
})).Return(nil)

// Specify call count
repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()    // Called once
repo.On("Update", mock.Anything, mock.Anything).Return(nil).Times(3)  // Called 3 times
repo.On("Update", mock.Anything, mock.Anything).Return(nil)           // Any number of times
```

### Mocking Internal Hooks

Some packages expose package-level function variables so tests can replace external calls, for example `newEmailSender` and `parseForgotTemplate` in `internal/services/mail_service.go`, or `openGormConnection` and `pingDBFn` in `internal/configs/database.go`. Such tests use an internal test package (`package services`, `package configs`; see `*_internal_test.go`) and restore the original value with `t.Cleanup`.

---

## Test Organization

### File Naming

```
service.go                   // Production code
service_test.go              // Unit tests (external package, e.g. services_test)
service_internal_test.go     // Tests that need unexported identifiers (same package)
```

### Test Package Organization

```
internal/
├── services/
│   ├── user_service.go
│   ├── user_service_test.go          # package services_test
│   ├── jwt_service.go
│   ├── jwt_service_test.go
│   └── jwt_service_internal_test.go  # package services
├── handlers/
│   ├── user_handler.go
│   └── user_handler_test.go
└── repositories/
    ├── user_repository.go
    └── user_repository_test.go
tests/
├── e2e/                              # Full-router tests
└── mocks/                            # Shared mocks
```

### Grouped Test Functions

```go
// Old style - multiple separate test functions
func TestGetProfile(t *testing.T) { }
func TestUpdateProfile(t *testing.T) { }

// Preferred style - grouped test functions
func TestUserService(t *testing.T) {
    t.Run("GetProfile", func(t *testing.T) { })
    t.Run("UpdateProfile", func(t *testing.T) { })
}
```

### Shared Test Fixtures

```go
// Helper function for test setup (in-memory SQLite, no cleanup needed)
func setupUserTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    require.NoError(t, db.AutoMigrate(&models.User{}))
    return db
}

// Usage in tests
func TestUserRepository(t *testing.T) {
    db := setupUserTestDB(t)
    repo := repositories.NewUserRepository(db)
    // ... test code
}
```

---

## Coverage Goals

### Target Coverage by Layer

| Layer | Target | Rationale |
|-------|--------|-----------|
| Handlers | 95%+ | Critical for API contracts |
| Services | 85%+ | Core business logic |
| Repositories | 90%+ | Data access is crucial |
| Middlewares | 85%+ | Security-related |
| Utils | 80%+ | General utilities |
| Models | 0% | Usually just data structures |

CI (`.github/workflows/build.yml`) runs the unit tests of every package except `cmd`, `docs`, and `tests`, and fails if total coverage is below **70%**. `internal/models`, `internal/routes`, and `internal/database/seeders` have no unit tests of their own (routes are covered by the e2e suite).

### Generate Coverage Report

```bash
# Makefile: core packages, writes coverage.out, coverage-summary.txt, coverage.html
make test-coverage

# Coverage for all packages
go test ./... -cover

# Detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage for a specific package
go test ./internal/services -cover

# Total coverage percentage (what CI checks)
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
```

### Measuring Coverage

```bash
# Show uncovered functions
go tool cover -func=coverage.out | grep 0$

# HTML report with coverage visualization
go tool cover -html=coverage.out -o coverage.html
```

---

## Common Test Patterns

### Table-Driven Middleware Tests

Many middleware tests define a table of cases with a mock setup function (see `internal/middlewares/auth_middleware_test.go`):

```go
tests := []struct {
    name               string
    authHeader         string
    mockSetup          func(*mocks.MockJWTService)
    expectedStatusCode int
    expectNext         bool
}{
    {
        name:               "missing Authorization header",
        authHeader:         "",
        mockSetup:          func(m *mocks.MockJWTService) {},
        expectedStatusCode: http.StatusUnauthorized,
        expectNext:         false,
    },
    // ...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        jwtService := new(mocks.MockJWTService)
        tt.mockSetup(jwtService)
        // build a router with middlewares.AuthMiddleware(jwtService), send the request,
        // and assert on the status code and whether the next handler ran
    })
}
```

### Parametrized Tests

```go
func TestValidateBirthday(t *testing.T) {
    tests := []struct {
        name     string
        birthday string
        valid    bool
    }{
        {"Valid date", "2000-01-01", true},
        {"Invalid month", "2000-13-01", false},
        {"Not a date", "tomorrow", false},
        {"Empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // validate tt.birthday with the valid_birthday rule and compare with tt.valid
        })
    }
}
```

### Error Testing

Services return `*apperror.AppError`; assert on its code and HTTP status rather than on message text alone.

```go
func TestErrorHandling(t *testing.T) {
    t.Run("Returns custom error", func(t *testing.T) {
        _, err := service.GetProfile(context.Background(), 9999)

        require.Error(t, err)
        var appErr *apperror.AppError
        require.True(t, errors.As(err, &appErr))
        assert.Equal(t, http.StatusNotFound, appErr.HttpStatusCode)
        assert.Equal(t, apperror.ErrNotFound, appErr.Code)
    })

    t.Run("Error contains message", func(t *testing.T) {
        _, err := service.GetProfile(context.Background(), 9999)

        assert.ErrorContains(t, err, "User not found")
    })
}
```

### Asynchronous Logging

`LogMiddleware` writes its log entry from a goroutine. Tests that assert on log output must wait for it (for example with `assert.Eventually`) instead of reading the buffer immediately, otherwise they are flaky.

---

## Performance Testing

### Benchmarking

```go
func BenchmarkHashPassword(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, _ = utils.HashPassword("SecurePassword123!")
    }
}

// Run benchmark
// go test -bench=. -benchtime=10s ./internal/shared/utils
```

The project does not currently contain benchmarks or load tests; add them next to the code they measure.

---

## Best Practices Summary

✅ **DO:**
- Group related tests under parent test functions
- Use `require` for critical assertions
- Use `assert` for value checks
- Mock external dependencies
- Test both success and error cases
- Use table-driven tests for multiple scenarios
- Keep tests independent and isolated
- Use meaningful test names
- Test edge cases and boundary conditions
- Maintain high test coverage

❌ **DON'T:**
- Test multiple things in one test
- Ignore test failures
- Write tests that depend on each other
- Skip error handling tests
- Use time-based assertions (flaky tests)
- Mock internal dependencies
- Have complex setup in tests
- Mix test levels (unit + integration)

---

Last Updated: September 30, 2026

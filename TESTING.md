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

- **Unit Tests** (80%): Fast, isolated, single function/method
- **Integration Tests** (15%): Multiple components, external services mocked
- **E2E Tests** (5%): Full user flows, real database

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
mockRepo.On("GetByID", uint(1)).Return(&User{}, nil)
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
assert.IsType(t, (*User)(nil), result)      // Check type
assert.Implements(t, (*Reader)(nil), obj)   // Implements interface
```

---

## Unit Testing

### Test Structure (AAA Pattern)

```go
func TestUserService(t *testing.T) {
    // Setup phase - creates all necessary objects
    mockRepo := new(mocks.MockUserRepository)
    service := services.NewUserService(mockRepo)

    t.Run("CreateUser - Success", func(t *testing.T) {
        // Arrange - set up test data and expectations
        user := &User{
            Email: "test@example.com",
            Name:  "Test User",
        }
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
}
```

### Handler Testing

```go
func TestUserHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("GetUser - Success", func(t *testing.T) {
        // Arrange
        mockService := new(mocks.MockUserService)
        expectedUser := &User{
            ID:    1,
            Email: "test@example.com",
            Name:  "Test User",
        }
        mockService.On("GetUser", uint(1)).Return(expectedUser, nil)

        handler := handlers.NewUserHandler(mockService)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("GET", "/api/v1/users/1", nil)
        c.Params = gin.Params{{Key: "id", Value: "1"}}

        // Act
        handler.GetUser(c)

        // Assert
        require.Equal(t, http.StatusOK, w.Code)
        var response map[string]interface{}
        require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
        assert.NotNil(t, response["data"])
    })

    t.Run("GetUser - Not Found", func(t *testing.T) {
        mockService := new(mocks.MockUserService)
        mockService.On("GetUser", uint(999)).Return(
            nil,
            apperror.NewNotFoundError("User not found"),
        )

        handler := handlers.NewUserHandler(mockService)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("GET", "/api/v1/users/999", nil)
        c.Params = gin.Params{{Key: "id", Value: "999"}}

        handler.GetUser(c)

        require.Equal(t, http.StatusNotFound, w.Code)
    })
}
```

### Repository Testing

```go
func TestUserRepository(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // Setup in-memory database
        db := setupTestDB()
        repo := repositories.NewUserRepository(db)

        // Arrange
        user := &User{
            Email: "test@example.com",
            Name:  "Test",
        }

        // Act
        result, err := repo.Create(user)

        // Assert
        require.NoError(t, err)
        assert.Greater(t, result.ID, uint(0))
        assert.Equal(t, user.Email, result.Email)
    })

    t.Run("GetByID", func(t *testing.T) {
        db := setupTestDB()
        repo := repositories.NewUserRepository(db)

        // Setup - create a user
        created := &User{Email: "test@example.com"}
        db.Create(created)

        // Act
        result, err := repo.GetByID(created.ID)

        // Assert
        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, created.Email, result.Email)
    })

    t.Run("GetByID - Not Found", func(t *testing.T) {
        db := setupTestDB()
        repo := repositories.NewUserRepository(db)

        // Act
        result, err := repo.GetByID(9999)

        // Assert
        assert.Error(t, err)
        assert.Nil(t, result)
    })
}
```

### Service Testing

```go
func TestAuthService(t *testing.T) {
    t.Run("Login - Success", func(t *testing.T) {
        // Arrange
        mockUserRepo := new(mocks.MockUserRepository)
        mockBcryptService := new(mocks.MockBcryptService)
        mockJWTService := new(mocks.MockJWTService)

        user := &User{
            ID:       1,
            Email:    "test@example.com",
            Password: "hashed_password",
        }

        mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
        mockBcryptService.On("CheckPasswordHash", "password", "hashed_password").
            Return(true)
        mockJWTService.On("GenerateAccessToken", uint(1)).
            Return(&JwtResult{Token: "jwt_token", ExpiresAt: 1000}, nil)

        service := services.NewAuthService(
            mockUserRepo,
            mockBcryptService,
            mockJWTService,
        )

        // Act
        result, err := service.Login("test@example.com", "password")

        // Assert
        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, "jwt_token", result.Token)
    })

    t.Run("Login - Invalid Credentials", func(t *testing.T) {
        mockUserRepo := new(mocks.MockUserRepository)
        mockBcryptService := new(mocks.MockBcryptService)
        mockJWTService := new(mocks.MockJWTService)

        user := &User{
            ID:       1,
            Email:    "test@example.com",
            Password: "hashed_password",
        }

        mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
        mockBcryptService.On("CheckPasswordHash", "wrong_password", "hashed_password").
            Return(false)

        service := services.NewAuthService(
            mockUserRepo,
            mockBcryptService,
            mockJWTService,
        )

        result, err := service.Login("test@example.com", "wrong_password")

        require.Error(t, err)
        assert.Nil(t, result)
    })
}
```

---

## Integration Testing

### Database Integration Tests

```go
func TestUserServiceIntegration(t *testing.T) {
    // Setup real database (in-memory for tests)
    db := setupTestDB()
    defer teardownTestDB(db)

    userRepo := repositories.NewUserRepository(db)
    bcryptService := services.NewBcryptService()
    userService := services.NewUserService(userRepo, bcryptService)

    t.Run("Create and Retrieve User", func(t *testing.T) {
        // Act 1: Create user
        created, err := userService.CreateUser(&CreateUserRequest{
            Email:    "test@example.com",
            Password: "SecurePassword123!",
            Name:     "Test User",
        })

        // Assert 1
        require.NoError(t, err)
        assert.Greater(t, created.ID, uint(0))
        assert.Equal(t, "test@example.com", created.Email)

        // Act 2: Retrieve user
        retrieved, err := userService.GetUser(created.ID)

        // Assert 2
        require.NoError(t, err)
        assert.Equal(t, created.Email, retrieved.Email)
    })
}
```

### Multiple Service Integration

```go
func TestAuthServiceIntegration(t *testing.T) {
    db := setupTestDB()
    defer teardownTestDB(db)

    // Setup services with real dependencies
    userRepo := repositories.NewUserRepository(db)
    refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
    bcryptService := services.NewBcryptService()
    jwtService := services.NewJWTService()
    refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)
    authService := services.NewAuthService(
        userRepo,
        bcryptService,
        jwtService,
        refreshTokenService,
    )

    t.Run("Complete Auth Flow", func(t *testing.T) {
        // Create user
        userService := services.NewUserService(userRepo, bcryptService)
        user, err := userService.CreateUser(&CreateUserRequest{
            Email:    "test@example.com",
            Password: "Password123!",
            Name:     "Test User",
        })
        require.NoError(t, err)

        // Login
        loginResult, err := authService.Login(
            "test@example.com",
            "Password123!",
        )
        require.NoError(t, err)
        assert.NotEmpty(t, loginResult.AccessToken.Token)
        assert.NotEmpty(t, loginResult.RefreshToken.Token)

        // Verify token
        claims, err := jwtService.ValidateToken(loginResult.AccessToken.Token)
        require.NoError(t, err)
        assert.Equal(t, user.ID, claims.ID)
    })
}
```

---

## Mocking Strategy

### Mocking External Dependencies

```go
// HTTP Client Mock
type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    args := m.Called(req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*http.Response), args.Error(1)
}

// Usage
mockClient := new(MockHTTPClient)
response := &http.Response{
    StatusCode: 200,
    Body:       io.NopCloser(strings.NewReader(`{"id":1}`)),
}
mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
    return req.Method == "GET" && strings.Contains(req.URL.Path, "/users/1")
})).Return(response, nil)
```

### Mocking with Arguments

```go
// Match specific arguments
mockRepo.On("GetByID", 1).Return(&User{ID: 1}, nil)

// Match argument by type
mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

// Match argument by function
mockRepo.On("Create", mock.MatchedBy(func(u *User) bool {
    return u.Email != "" && len(u.Name) > 0
})).Return(nil)

// Multiple calls with different returns
mockService.On("Process", "input1").Return("output1", nil)
mockService.On("Process", "input2").Return("output2", nil)

// Specify call count
mockRepo.On("Create", mock.Anything).Return(nil).Once()      // Called once
mockRepo.On("Update", mock.Anything).Return(nil).Times(3)   // Called 3 times
mockRepo.On("Delete", mock.Anything).Return(nil)            // Called any number of times
```

### Complete Mock Example

```go
func TestPaymentService(t *testing.T) {
    t.Run("ProcessPayment - Success", func(t *testing.T) {
        // Create mocks
        mockGateway := new(mocks.MockPaymentGateway)
        mockEmailService := new(mocks.MockEmailService)
        mockLogger := new(mocks.MockLogger)

        // Setup expectations
        mockGateway.On(
            "Charge",
            mock.MatchedBy(func(req *ChargeRequest) bool {
                return req.Amount > 0 && req.CardToken != ""
            }),
        ).Return(&ChargeResponse{
            TransactionID: "txn_123",
            Status:        "success",
        }, nil)

        mockEmailService.On(
            "Send",
            mock.MatchedBy(func(email *Email) bool {
                return strings.Contains(email.Subject, "Payment")
            }),
        ).Return(nil)

        mockLogger.On("Info", mock.MatchedBy(func(msg string) bool {
            return strings.Contains(msg, "Payment processed")
        })).Return()

        // Create service with mocks
        service := services.NewPaymentService(
            mockGateway,
            mockEmailService,
            mockLogger,
        )

        // Execute
        result, err := service.ProcessPayment(&PaymentRequest{
            Amount:    100.00,
            CardToken: "tok_visa",
            Email:     "customer@example.com",
        })

        // Assert
        require.NoError(t, err)
        assert.Equal(t, "txn_123", result.TransactionID)
        assert.Equal(t, "success", result.Status)

        // Verify all mocks were called as expected
        mockGateway.AssertExpectations(t)
        mockEmailService.AssertExpectations(t)
        mockLogger.AssertExpectations(t)
    })
}
```

---

## Test Organization

### File Naming

```
service.go              // Production code
service_test.go         // Unit tests
service_integration_test.go  // Integration tests
```

### Test Package Organization

```
internal/
├── services/
│   ├── user_service.go
│   ├── user_service_test.go         # Unit tests
│   ├── user_service_integration_test.go
│   ├── auth_service.go
│   └── auth_service_test.go
├── handlers/
│   ├── user_handler.go
│   └── user_handler_test.go
└── repositories/
    ├── user_repository.go
    └── user_repository_test.go
```

### Grouped Test Functions

```go
// Old style - multiple separate test functions
func TestCreateUser(t *testing.T) { }
func TestGetUser(t *testing.T) { }
func TestUpdateUser(t *testing.T) { }

// New style - grouped test functions
func TestUserService(t *testing.T) {
    t.Run("CreateUser", func(t *testing.T) { })
    t.Run("GetUser", func(t *testing.T) { })
    t.Run("UpdateUser", func(t *testing.T) { })
}
```

### Shared Test Fixtures

```go
// Helper function for test setup
func setupTestDB() *gorm.DB {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&User{})
    return db
}

// Helper function for test cleanup
func teardownTestDB(db *gorm.DB) {
    sqlDB, _ := db.DB()
    sqlDB.Close()
}

// Usage in tests
func TestUserRepository(t *testing.T) {
    db := setupTestDB()
    defer teardownTestDB(db)
    
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

### Generate Coverage Report

```bash
# Generate coverage for all packages
go test ./... -cover

# Generate detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage for specific package
go test ./internal/services -cover

# Fail if coverage below threshold
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
```

### Measuring Coverage

```bash
# Show uncovered lines
go tool cover -func=coverage.out | grep 0$

# HTML report with coverage visualization
go tool cover -html=coverage.out -o coverage.html
```

---

## Common Test Patterns

### Before/After Pattern

```go
func setupTest(t *testing.T) *TestContext {
    // Setup database
    db := setupTestDB()
    
    // Setup repositories
    userRepo := repositories.NewUserRepository(db)
    
    // Setup services
    userService := services.NewUserService(userRepo)
    
    return &TestContext{
        db:          db,
        userRepo:    userRepo,
        userService: userService,
    }
}

func teardownTest(t *testing.T, ctx *TestContext) {
    teardownTestDB(ctx.db)
}

func TestUserService(t *testing.T) {
    ctx := setupTest(t)
    defer teardownTest(t, ctx)

    t.Run("CreateUser", func(t *testing.T) {
        // Use ctx.userService, ctx.userRepo, etc.
    })
}
```

### Parametrized Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        isValid bool
    }{
        {"Valid email", "user@example.com", true},
        {"Missing domain", "user@", false},
        {"No at symbol", "userexample.com", false},
        {"Empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidateEmail(tt.email)
            assert.Equal(t, tt.isValid, result)
        })
    }
}
```

### Error Testing

```go
func TestErrorHandling(t *testing.T) {
    t.Run("Returns custom error", func(t *testing.T) {
        _, err := service.GetUser(9999)
        
        require.Error(t, err)
        var appErr *apperror.AppError
        assert.True(t, errors.As(err, &appErr))
        assert.Equal(t, http.StatusNotFound, appErr.HttpStatusCode)
    })

    t.Run("Error contains message", func(t *testing.T) {
        _, err := service.GetUser(9999)
        
        assert.ErrorContains(t, err, "not found")
    })
}
```

---

## Performance Testing

### Benchmarking

```go
func BenchmarkHashPassword(b *testing.B) {
    password := "SecurePassword123!"
    for i := 0; i < b.N; i++ {
        bcrypt.GenerateFromPassword([]byte(password), 10)
    }
}

// Run benchmark
// go test -bench=. -benchtime=10s
```

### Load Testing

```go
func TestConcurrentRequests(t *testing.T) {
    const numGoroutines = 100
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            user, err := service.GetUser(uint(id))
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("Error: %v", err)
    }
}
```

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

Last Updated: November 21, 2025

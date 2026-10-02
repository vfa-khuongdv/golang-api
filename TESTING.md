# Testing Standards

Testing rules for Golang CMS using [Testify](https://github.com/stretchr/testify) (`assert`, `require`, `mock`, `suite`).

## Rules

- **TDD**: write the failing test first (red), then the minimum code to pass (green), then refactor. For a bug, first write a test that reproduces it.
- Arrange-Act-Assert; group cases with `t.Run("Name - Case", ...)`; table-driven tests for many inputs.
- `require` for checks that must stop the test (errors, nil), `assert` for value checks.
- Test success and failure paths and boundaries. Keep tests independent; avoid time-based assertions.
- Mock external dependencies (repositories, services, SMTP), not the code under test.

## Layout

```
internal/**/*_test.go, pkg/**/*_test.go   unit tests next to the code
  package x_test                          black-box tests (default)
  *_internal_test.go (package x)          tests that need unexported identifiers
tests/e2e                                 full-router tests (in-memory SQLite)
tests/mocks                               shared hand-written Testify mocks
```

| Kind | How |
|------|-----|
| Service | Mock repositories/services from `tests/mocks` |
| Handler | `gin.CreateTestContext`, mocked services, call the handler method directly |
| Repository | In-memory SQLite (`gorm.Open(sqlite.Open(":memory:"))`) |
| Middleware | Table-driven with a Gin router and mocked services |
| End-to-end | Real router via `setupTestRouter()` in `tests/e2e/setup_test.go` |

Mocks in `tests/mocks`: `MockUserRepository`, `MockRefreshTokenRepository`, `MockSettingRepository`, `MockAuthService`, `MockUserService`, `MockRefreshTokenService`, `MockJWTService`, `MockMailerService`, `MockDB`/`MockTx`. Add one for every new repository or service interface.

## Commands

```bash
make test             # unit tests (excludes cmd, docs, tests)
make test-e2e         # tests/e2e
make test-coverage    # coverage.out, coverage-summary.txt, coverage.html (core packages)
make watch-test       # re-run on change (needs reflex)
go test ./internal/services -run TestUserService -v
go test ./... -race
```

## Coverage

Targets: handlers 95%, services 85%, repositories 90%, middlewares 85%, utils 80%. CI runs the unit tests of every package except `cmd`, `docs` and `tests` and fails below **70%** total. `internal/models`, `internal/routes` and the seeders have no unit tests (routes are covered by e2e).

## Patterns

### Service

Every service and repository method takes a `context.Context` first, so mocks match it with `mock.Anything`.

```go
package services_test

func TestUserService(t *testing.T) {
    t.Run("GetProfile - Success", func(t *testing.T) {
        repo := new(mocks.MockUserRepository)
        service := services.NewUserService(repo, new(mocks.MockMailerService), new(mocks.MockRefreshTokenService))
        user := &models.User{ID: 1, Email: "test@example.com"}
        repo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

        result, err := service.GetProfile(context.Background(), 1)

        require.NoError(t, err)
        assert.Equal(t, user, result)
        repo.AssertExpectations(t)
    })
}
```

Assert on the error code rather than the message:

```go
appErr, ok := apperror.ToAppError(err)
require.True(t, ok)
assert.Equal(t, apperror.ErrInvalidPassword, appErr.Code)
```

Service tests can also use `suite.Suite` (see `internal/services/user_service_test.go`): build mocks in `SetupTest`, assert expectations in `TearDownTest`. `NewAuthService(userRepo, refreshTokenService, jwtService)` takes the user repository, refresh token service and JWT service.

### Handler

Handlers are concrete structs. Call `utils.InitValidator()` so the custom binding rules (`password_complexity`, `not_blank`, `valid_birthday`) exist, and set the user like `AuthMiddleware` does.

```go
package handlers_test

func TestGetProfile(t *testing.T) {
    gin.SetMode(gin.TestMode)
    utils.InitValidator()

    t.Run("Not Found", func(t *testing.T) {
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

Responses are plain JSON: `RespondWithOK` writes the body you pass (no `data` envelope); errors are `{"code","message"}` plus `fields` for validation errors. Use the real HTTP method and path of the route.

### Repository

```go
func setupUserTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    require.NoError(t, db.AutoMigrate(&models.User{})) // only the models the test needs
    return db
}
```

### End-to-end

`setupTestRouter()` sets `JWT_KEY` and `SETTINGS_ENCRYPTION_KEY`, opens a shared in-memory SQLite (one connection), migrates `User`, `RefreshToken` and `Setting`, calls `utils.InitValidator()`, and returns the `*gin.Engine` and `*gorm.DB`. Create data directly with `db`, then send requests through `router.ServeHTTP`:

```go
router, db := setupTestRouter()
hashed, _ := utils.HashPassword("password123")
require.NoError(t, db.Create(&models.User{Name: "Test", Email: "t@example.com", Password: hashed, Gender: 1}).Error)

body, _ := json.Marshal(map[string]string{"email": "t@example.com", "password": "password123"})
w := httptest.NewRecorder()
req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
req.Header.Set("Content-Type", "application/json")
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
```

Current suites cover login, logout, refresh token, forgot/reset password, change password, get/update profile, empty bodies, rate limiting, health and version.

### Replacing external calls

Some packages expose function variables so tests can swap external calls (`newEmailSender`, `parseForgotTemplate` in `internal/services/mail_service.go`; `openGormConnection`, `pingDBFn` in `internal/configs/database.go`). Use an internal test package and restore the original with `t.Cleanup`.

### Asynchronous logging

`LogMiddleware` writes its entry from a goroutine. Wait for it with `assert.Eventually` before asserting on log output, otherwise the test is flaky.

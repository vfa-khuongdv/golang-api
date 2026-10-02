# Development Guidelines

Project-specific standards for Golang CMS. Setup, environment variables, endpoints and `make` targets are in [README.md](README.md); testing details are in [TESTING.md](TESTING.md); logging is in [docs/logging-standards.md](docs/logging-standards.md).

## Architecture

Clean architecture: `handlers → services → repositories → database`. Dependencies are wired by hand in `SetupRouter` (`internal/routes/routes.go`): repositories → services → handlers.

```
cmd/{server,seeder}   entry points
internal/
  configs/        env config + DB connection
  database/       migrations/ (SQL, golang-migrate), seeders/
  handlers/ services/ repositories/ models/ middlewares/
  routes/         router setup and wiring
  shared/{constants,dto,utils}
pkg/{apperror,logger,mailer,migrator}
tests/{e2e,mocks}                     unit tests sit next to the code (*_test.go)
docs/                                 swagger.json, logging-standards.md
```

| Layer | Responsibility |
|-------|----------------|
| Models | GORM + JSON tags, `TableName()`, no logic. Hide internals with `json:"-"` (`Password`, `ResetToken`) |
| DTOs (`shared/dto`) | Request/response types with `binding` tags (`required`, `email`, `password_complexity`, ...) |
| Repositories | DB access only, behind interfaces |
| Services | Business rules (lockout, token expiry, password rules); return `*apperror.AppError` |
| Handlers | Bind + validate, call a service, respond |
| Middlewares | Request ID, CORS, request logging, JWT auth, rate limiting; `gin.Recovery()` handles panics |

Empty request bodies are rejected centrally by `utils.TranslateValidationErrors`, not by a middleware.

## Conventions

- JSON field names are snake_case (`created_at`, `expires_at`).
- New constants use CamelCase (`MaxFailedAttempts`, `SettingMailHost`). A few older ones are upper snake case (`LIMIT`, `MAX_BODY_SIZE`); leave them.
- Interfaces have no `I` prefix and stay small. Repository and service constructors return the interface; handler constructors return the concrete struct (tests call methods directly). `UserRepository` and `UserService` are larger than ideal; keep new interfaces small.
- Every service and repository method takes `context.Context` first.
- Verbs: `Get`, `Create`, `Update`, `Delete`; `Is`/`Has` for booleans.
- Comments explain why, not what (see the ordering comment in `AuthService.RefreshToken`).

```go
type SettingRepository interface {
    GetValues(ctx context.Context, keys ...string) (map[string]string, error)
}
func NewSettingRepository(db *gorm.DB) SettingRepository { return &settingRepositoryImpl{db: db} }
```

## Error Handling

`pkg/apperror`:

```go
type AppError struct {
    HttpStatusCode int    `json:"-"`
    Code           int    `json:"code"`    // see pkg/apperror/codes.go
    Message        string `json:"message"`
    Err            error  `json:"-"`       // underlying cause
}
```

- Codes: general 1000s, database 2000s, authentication 3000s, common/cache 4000s.
- Create with `apperror.New(status, code, msg)` or `apperror.Wrap(status, code, msg, err)`, or use the factories: `NewBadRequestError` (400), `NewUnauthorizedError` (401), `NewForbiddenError` (403), `NewNotFoundError` (404), `NewConflictError` (409), `NewInternalServerError` (500), `NewInvalidPasswordError` (400), `NewAccountLockedError` (429), `NewDB{Query,Insert,Update,Delete}Error` (500).
- Validation failures use `apperror.ValidationError` (`NewValidationError(msg, []FieldError{...})`), produced in handlers by `utils.TranslateValidationErrors`.
- `utils.RespondWithError` maps `ValidationError` → 400 with `fields`, `AppError` → its status, anything else → 500 "Internal server error". `utils.RespondWithOK(ctx, status, body)` writes `body` as-is (no envelope).
- Never return raw DB errors to clients.

Handler pattern:

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
    if _, err = handler.userService.ChangePassword(ctx.Request.Context(), userId, &input); err != nil {
        logger.WithEvent(ctx.Request.Context(), logger.EventPasswordChangeFailed).
            Errorf("Change password failed for user %d: %v", userId, err)
        utils.RespondWithError(ctx, err)
        return
    }
    utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
}
```

Service pattern:

```go
func (service *userServiceImpl) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
    user, err := service.repo.GetByID(ctx, userID)
    if err != nil {
        return nil, apperror.NewNotFoundError("User not found")
    }
    return user, nil
}
```

## Testing (TDD)

Write the test before the code, for every feature and bug fix:

1. **Red**: write a failing test and confirm it fails for the expected reason (for a bug, a test that reproduces it)
2. **Green**: write the minimum code to pass
3. **Refactor** with the tests green

Do not write production code without a failing test first. Patterns, mocks, e2e setup and coverage targets are in [TESTING.md](TESTING.md). CI fails below 70% total coverage of the unit-tested packages (everything except `cmd`, `docs`, `tests`).

## Git Workflow

- Branches: `feat/…`, `fix/…`, `chore/…`, `docs/…`; open PRs against `main`.
- Conventional commits: `<type>(<scope>): <subject>` with types `feat`, `fix`, `refactor`, `test`, `docs`, `style`, `chore`.
- One feature or fix per PR; include tests; update docs when the API or configuration changes.
- CI (`.github/workflows/build.yml`) runs on every PR and push to `main`, with these jobs in parallel:
  - **Lint**: `go mod tidy -diff`, then golangci-lint (including gofmt and gosec, see `.golangci.yml`).
  - **Tests & Coverage**: unit and e2e tests with `-race`, and the 70% coverage gate.
  - **Security**: govulncheck (vulnerabilities in code the app calls) and gitleaks over the git history.
  - **Integration (MySQL)**: `scripts/smoke-test.sh` applies the migrations on MySQL 8.0, seeds, and calls the API (login, permissions, lockout). The other tests use SQLite, so this catches MySQL-only problems. Run it locally against an empty database with the `DB_*`, `JWT_KEY` and `SETTINGS_ENCRYPTION_KEY` variables set.
  - **Docker Build**: builds the image without pushing it.

  `make pre-push` (fmt, vet, lint, test) covers the lint and unit-test part locally. A gitleaks pre-commit hook is configured in `.pre-commit-config.yaml`.

## API Documentation

`docs/swagger.json` is maintained by hand (no code annotations). When you add, change or remove an endpoint, update it together with `internal/routes/routes.go` and the endpoint list in `README.md`. Swagger UI is only served when `STAGE` is not `prod`.

## Database

- Schema changes are SQL migrations in `internal/database/migrations` (`*.up.sql` and `*.down.sql`), applied by golang-migrate on startup when `RUN_MIGRATE=true` (run the server from the repository root).
- Use transactions for multi-step writes (`BeginTx` in the repositories; refresh-token rotation is an example).
- Paginate with `dto.Pagination` and `utils.ParsePageAndLimit` (default page size `constants.LIMIT`).
- Pool size: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS` (`internal/configs/database.go`).

## Security

- **Access JWT**: 1 hour, `access` scope, HMAC-signed only. **Refresh token**: 60 random characters, 30 days, stored in the database, rotated on refresh, deleted on logout. **Reset token**: 1 hour, only its hash is stored.
- **Passwords**: bcrypt; `password_complexity` requires 8+ characters with upper, lower, digit and special character. 5 failed logins lock the account for 15 minutes (`services.MaxFailedAttempts`, `LockoutDurationMinutes`). The four public auth endpoints are limited to 10 requests/minute per IP. Never reveal whether an email exists.
- **Secrets**: `JWT_KEY` and `SETTINGS_ENCRYPTION_KEY` (32+ characters) come from the environment. Secret `settings` rows (`mail.password`) are AES-256-GCM encrypted (`utils.EncryptSecret`), set through `PUT /api/v1/settings`. Mail and frontend settings live in the `settings` table, not env vars.
- **Input**: validate with `binding` tags on DTOs; use GORM parameterized queries; never log raw sensitive values (`utils.MaskWithPrefix`; the log middleware masks bodies and headers).
- **CORS**: `CORS_ALLOWED_ORIGINS` (exact origins, default `http://localhost:5173`); credentials are allowed, so never use `*` in production.
- **Proxies**: `TRUSTED_PROXIES` defaults to `0.0.0.0/0` so the client IP (rate limiter key) is read from `X-Forwarded-For` behind an ALB or reverse proxy without configuring its IP. The trade-off is that a client can spoof that header and dodge the per-IP rate limit; set the proxy CIDR to prevent it, or set an empty value when the app is exposed directly with no proxy. Trusting nothing behind a proxy makes all clients share the proxy's IP and throttles every user together.

## Deployment

The `Dockerfile` builds a static binary (`golang:1.27-alpine`) and runs it from `alpine:3.21` as a non-root user on port 3000. It copies `internal/` (for migrations) but not `docs/`, so Swagger is unavailable in the image unless you add it. There is no `HEALTHCHECK`; probe `GET /healthz` if you need one. `docker-compose.yml` only provides local dependencies (MySQL, phpMyAdmin, Mailpit), not the app.

## Code Review Checklist

- [ ] Failing test written first; tests pass
- [ ] Errors use `apperror`; no raw DB errors returned
- [ ] `context.Context` passed through all layers
- [ ] Layers stay separate; dependencies injected
- [ ] No hardcoded config; no sensitive data logged
- [ ] Migration added for schema changes
- [ ] `docs/swagger.json` and README updated for API changes

Last Updated: September 30, 2026

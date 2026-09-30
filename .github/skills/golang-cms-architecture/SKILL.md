---
name: golang-cms-architecture
description: Development guidelines for golang-cms, a Go CMS REST API with clean architecture (handlers → services → repositories → models). Use when implementing features, writing tests, or modifying the golang-cms codebase. Covers authentication, error handling, TDD with testify, logging, and conventions.
license: MIT
metadata:
  author: golang-cms
  version: "1.1"
---

# Golang CMS Architecture & Development Guide

Go 1.27+, Gin, GORM, MySQL 8 (SQLite in-memory for tests), testify. Commands, setup, env vars and the endpoint list are in `README.md` and the `Makefile`; do not duplicate them here.

## Structure

```
cmd/{server,seeder,encrypt-setting}   entry points
internal/
  configs/        env config + DB connection
  database/       migrations/ (SQL, golang-migrate), seeders/
  handlers/ services/ repositories/ models/ middlewares/
  routes/         SetupRouter: wires repos → services → handlers
  shared/{constants,dto,utils}
pkg/{apperror,logger,mailer,migrator}
tests/{e2e,mocks}                     unit tests sit next to the code
docs/                                 swagger.json (hand-written), logging-standards.md
```

## Layers

| Layer | Responsibility |
|-------|----------------|
| Handlers | Bind + validate request, call service, respond |
| Services | Business rules, orchestrate repositories |
| Repositories | GORM queries behind interfaces |
| Models | GORM/JSON tags, no logic |
| DTOs (`shared/dto`) | Inputs with `binding` tags |

- Depend on interfaces; constructors of repos/services return interfaces, handlers return the concrete struct.
- Every service and repository method takes `context.Context` first. Handlers pass `ctx.Request.Context()`; repos use `db.WithContext(ctx)`.
- Keep interfaces small (1-3 methods).

## Conventions

- Files/packages lowercase; `Get/Create/Update/Delete` verbs; `New` constructors.
- JSON tags snake_case. New constants in CamelCase (`MaxFailedAttempts`).
- Hide internal fields with `json:"-"` (password, reset token).

## Errors and responses

```go
// Handler: bind, translate binding errors, call service, respond
var input dto.ChangePasswordInput
if err := ctx.ShouldBindJSON(&input); err != nil {
    utils.RespondWithError(ctx, utils.TranslateValidationErrors(err, input))
    return
}
if _, err := svc.ChangePassword(ctx.Request.Context(), userID, &input); err != nil {
    utils.RespondWithError(ctx, err)
    return
}
utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Change password successfully"})
```

- Services return `*apperror.AppError`: `NewBadRequestError`, `NewUnauthorizedError`, `NewNotFoundError`, `NewConflictError`, `NewInternalServerError`, `NewDB*Error`, `NewAccountLockedError`, or `apperror.Wrap(status, code, msg, err)` to keep the cause. Codes are in `pkg/apperror/codes.go`.
- `RespondWithError` handles `ValidationError` (400 + `fields`), `AppError` (its status), and anything else (500).
- `RespondWithOK(ctx, status, body)` writes `body` as-is, no envelope. Errors are `{"code","message"}`.
- Never return raw DB errors to clients.

## TDD (required)

Write the failing test first (red), make it pass with minimal code (green), then refactor. Applies to every layer and every bug fix (reproduce the bug with a failing test first).

- Testify `assert`/`require`/`mock`; group with `t.Run("Name - Case", ...)`; Arrange-Act-Assert.
- Mocks live in `tests/mocks` (hand-written); add one for each new interface.
- Repository tests: in-memory SQLite. Handler tests: `gin.CreateTestContext`, `utils.InitValidator()`, `c.Set("UserID", uint(1))`. Flows: `tests/e2e` via `setupTestRouter()`.
- Coverage targets: handlers 95%, services 85%, repos 90%, middlewares 85%, utils 80%; CI fails below 70% total.

Templates for each layer: `references/templates.md`.

## Adding a feature

1. Model (+ SQL migration `*.up.sql`/`*.down.sql` in `internal/database/migrations`)
2. Repository: failing test → interface + GORM impl
3. Service: failing test with mocked repos → business logic
4. DTO + handler: failing test → bind, call, respond
5. Wire in `routes.go` (`public` group is rate limited; `authenticated` needs a JWT)
6. e2e test in `tests/e2e`; add mocks for new interfaces
7. Update `docs/swagger.json` and the README endpoint list

## Security and auth (do not regress)

- Access JWT: 1 hour, `access` scope, HMAC only. Refresh token: 60 random chars, 30 days, stored in DB, rotated on refresh, deleted on logout. Reset token: 1 hour, only its hash is stored.
- 5 failed logins lock the account for 15 minutes; public auth routes are limited to 10 req/min per IP.
- Secret settings (`mail_password`) are AES-256-GCM encrypted with `SETTINGS_ENCRYPTION_KEY` (`make encrypt-setting`). Mail/frontend settings live in the `settings` table, not env vars.
- Never log sensitive values; see `docs/logging-standards.md`.

## Logging

- Request-scoped: `logger.WithContext(ctx)` (adds `request_id`) or `logger.WithEvent(ctx, logger.EventX)` (adds `event`). Plain `logger.Infof` only for startup/seeders.
- Add new event names as constants in `pkg/logger/logger.go`.

## Don't

Ignore errors, use globals, mix layers, hardcode config, store plaintext passwords, log sensitive data, write code before a failing test.

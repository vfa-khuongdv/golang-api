# Logging Standards

**Format:** JSON → stdout (12 Factor App). Every line is one JSON object.

---

## 1. Standard Fields

```json
{"level":"info","message":"Login successful for user ID 42","service":"golang-cms","env":"dev","version":"1.0.0","time":"2026-06-15T10:00:00Z","request_id":"abc","event":"login_success","latency_ms":45}
```

| Field | Always | Description |
|-------|--------|-------------|
| `level` | ✓ | `debug`, `info`, `warn`, `error`, `fatal` |
| `message` | ✓ | Human-readable description |
| `time` | ✓ | ISO8601 timestamp |
| `service` | ✓ | From `APP_SERVICE` (default `golang-cms`), set via `logger.Init()` |
| `env` | - | From `STAGE` (e.g. `local`, `dev`, `prod`; default `dev`); omitted when empty |
| `version` | ✓ | From `APP_VERSION` (default `dev`), set via `logger.Init()` |
| `request_id` | - | `X-Request-ID` header value or a generated UUID, added by `RequestIDMiddleware` |
| `event` | - | Event type for Kibana filtering (`logger.WithEvent`) |
| `latency_ms` | - | Processing time in milliseconds (login and token refresh events) |

`service`, `env`, and `version` are the defaults passed to `logger.Init()`. `time` is the standard logrus RFC 3339 timestamp and `message` is the logrus message field.

**Level usage:**
- `info` — key events: login, create, update
- `warn` — expected failures: wrong password, invalid token
- `error` — unexpected failures: DB errors, token generation failure

**Request-scoped logging:** use `logger.WithContext(ctx)` (adds `request_id`) or `logger.WithEvent(ctx, event)` (adds `request_id` and `event`). Use the plain `logger.Infof()` helpers only for startup, seeders, and other code without a request.

### HTTP access log

`LogMiddleware` writes one entry per request with the message `HTTP request completed` and these fields:

| Field | Description |
|-------|-------------|
| `request_id` | Request ID |
| `method`, `url` | HTTP method and URL |
| `status_code` | Response status code |
| `latency` | Processing time, formatted as `"<n> (ms)"` |
| `header` | Request headers (sensitive headers masked) |
| `request` | Query parameters, or the JSON body for `POST`/`PUT`/`PATCH` (sensitive fields masked, at most 64 KB) |
| `response` | Response body for status `>= 400`; `<not_log>` for successful responses |

The level is `info` below 400, `warn` for 4xx, and `error` for 5xx. The entry is written from a goroutine so it never blocks the response.

---

## 2. Events

Use the `event` field to filter logs in Kibana/Grafana.

```
login_attempt          → info   User attempts login
login_success          → info   Login succeeds (with latency_ms)
login_failed           → warn   Unknown email, locked account, or wrong password
                        error  Token generation/storage or handler-level failures
token_refresh          → info   Token refresh starts
token_refresh_success  → info   Token refresh succeeds (with latency_ms)
token_refresh_failed   → warn   Invalid/expired/mismatched tokens
                        error  Handler-level or token generation failures
password_reset_request → warn   Forgot-password for an unknown email
                        error  DB or mail failures while requesting a reset
password_reset         → error  Reset password fails (success is not logged with this event)
password_change_failed → error  Change password fails
profile_get            → info   Profile viewed
                        error  Profile lookup fails
profile_update_failed  → error  Profile update fails
logout                 → info   User logs out (refresh tokens revoked)
                        error  Logout fails
```

The constants `password_change` and `profile_update` exist in `pkg/logger/logger.go` but are not emitted yet; `password_reset` is currently only used for failures. The constants are defined in `pkg/logger/logger.go` (`logger.EventLoginFailed`, ...). Add new events there instead of using string literals.

Example Kibana query: `event: login_failed AND latency_ms: > 1000`

---

## 3. Sensitive Data

**Sensitive values are masked before logging.** Two helpers are used:

- `utils.MaskWithPrefix(value, 4)` keeps the first 4 characters and replaces the rest with `*****` (values of 4 characters or fewer keep only the first). It is used for emails in service logs, for headers, and for query parameters.
- `utils.CensorSensitiveData` masks the matching fields of JSON request/response bodies, keeping only the first 2 characters.

Field names masked in bodies and query parameters (case-insensitive, see `sensitiveKeys` in `internal/middlewares/log_middleware.go`): `password`, `new_password`, `old_password`, `confirm_password`, `token`, `access_token`, `refresh_token`, `email`, `phone`, `address`, `otp`, `totp`, `mfa_code`, `mfa_secret`, `verification_code`, `secret`, `api-key`, `credit_card`, `debit_card`, `cvv`, `ccv`, `ssn`, `social_security_number`, `bank_account`, `bank_account_number`, `session`, `session_id`, `sessionid`, and `sid`. Any field whose name contains `password`, `secret`, or `token` (e.g. `mail_password`) is masked too.

```
Service log (MaskWithPrefix, 4)
email           → "user@example.com"  → "user*****"

JSON body (CensorSensitiveData, 2)
password        → "MyP@ssw0rd!"       → "My*****"
email           → "user@example.com"  → "us*****"
refresh_token   → "abc123..."         → "ab*****"
```

Headers (first 4 characters of the value kept):
```
Authorization: Bearer eyJhbGci... → Bearer eyJh*****
Cookie: session_id=abc123         → session_id=abc1*****
```

Other masked headers: `Set-Cookie`, `X-Api-Key`, `X-Auth-Token`, `Proxy-Authorization`, `X-Forwarded-For`, `X-Real-Ip`, `Forwarded`, `True-Client-Ip`, and the CSRF/XSRF token headers.

Never log raw passwords, tokens, or reset links yourself; log user IDs or masked values instead.

---

## 4. Configuration

```go
// cmd/server/main.go
logger.Init(logger.LogConfig{
    ServiceName: cfg.App.ServiceName, // APP_SERVICE
    Stage:       cfg.Server.Stage,    // STAGE
    Version:     cfg.App.Version,     // APP_VERSION
})
```

Output goes to stdout as JSON (`logrus.JSONFormatter`). Set `APP_SERVICE`, `STAGE`, and `APP_VERSION` in the environment (see the README) instead of passing build flags.

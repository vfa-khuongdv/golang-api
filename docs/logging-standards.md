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
| `service` | ✓ | Set via `logger.Init()` |
| `env` | ✓ | `dev`, `staging`, `prod` |
| `version` | - | Set via ldflags: `-X main.appVersion=1.0.0` |
| `request_id` | - | UUID from middleware |
| `event` | - | Event type for Kibana filtering |
| `latency_ms` | - | Processing time in milliseconds |

**Level usage:**
- `info` — key events: login, create, update
- `warn` — expected failures: wrong password, invalid token
- `error` — unexpected failures: DB errors, token generation failure

---

## 2. Events

Use the `event` field to filter logs in Kibana/Grafana.

```
login_attempt        → info   User attempts login
login_success        → info   Login succeeds
login_failed         → warn   Wrong email/password
token_refresh        → info   Token refresh starts
token_refresh_success → info  Token refresh succeeds
token_refresh_failed → warn   Invalid/expired token
password_reset_request → warn Password reset requested
password_reset       → info   Password reset succeeds
password_change      → info   Password changed
password_change_failed → warn Password change fails
profile_get          → info   Profile viewed
profile_update       → info   Profile updated
profile_update_failed → warn  Profile update fails
```

Example Kibana query: `event: login_failed AND latency_ms: > 1000`

---

## 3. Sensitive Data

**All sensitive fields are masked before logging.** Strategy: keep first 4 chars, replace rest with `*****`.

```
password    → "MyP@ssw0rd!" → "MyP@*****"
email       → "user@example.com" → "user*****"
token       → "eyJhbGci..." → "eyJh*****"
credit_card → "4111-1111-1111-1111" → "4111*****"
phone       → "0912345678" → "0912*****"
```

Headers:
```
Authorization: Bearer eyJhbGci... → Bearer eyJh*****
Cookie: session_id=abc123         → session_id=abc1*****
```

See [log-security-examples.md](log-security-examples.md) for full reference.

---

## 4. Configuration

```go
// main.go
logger.Init(logger.LogConfig{
    ServiceName: "golang-cms",
    Stage:       cfg.Server.Stage,
    Version:     appVersion,  // -ldflags="-X main.appVersion=1.0.0"
})
```

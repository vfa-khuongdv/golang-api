# Login Flow Documentation

## Simple Login Flow

```
Client                       Server
  │                           │
  ├─ POST /api/auth/login ───>│
  │  { email, password }      │
  │                           │
  │                           ├─ Find user by email
  │                           │
  │                           ├─ Verify password
  │                           │
  │                           ├─ Generate tokens
  │                           │
  │<─ 200 OK ─────────────────┤
  │  { accessToken,           │
  │    refreshToken }         │
  │                           │
  ├─ Store tokens             │
  │                           │
```

## API Endpoints

- `POST /api/auth/login` - Login with email and password
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/mfa/verify` - Verify MFA code (if MFA enabled)

## Login Request

```json
POST /api/auth/login
{
  "email": "user@example.com",
  "password": "password123"
}
```

## Login Response (Without MFA)

```json
{
  "accessToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-16T12:30:00Z"
  },
  "refreshToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-23T11:15:00Z"
  }
}
```

## Login Response (With MFA Enabled)

```json
{
  "mfa_required": true,
  "temporary_token": "eyJhbGciOiJIUzI1NiIs...",
  "message": "MFA code required"
}
```

## MFA Verification Request

```json
POST /api/auth/mfa/verify
{
  "code": "123456",
  "temporary_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## MFA Verification Response

```json
{
  "accessToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-16T12:30:00Z"
  },
  "refreshToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-23T11:15:00Z"
  }
}
```

## Refresh Token Request

```json
POST /api/auth/refresh
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## Refresh Token Response

```json
{
  "accessToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-16T12:30:00Z"
  },
  "refreshToken": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expiresAt": "2025-11-23T11:15:00Z"
  }
}
```

## Using Access Token

Add the access token to the Authorization header:

```bash
curl -X GET http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer <accessToken>"
```

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 400 | VALIDATION_ERROR | Invalid input |
| 401 | INVALID_PASSWORD | Wrong password |
| 401 | INVALID_TOKEN | Invalid/expired token |
| 404 | USER_NOT_FOUND | User doesn't exist |
| 500 | INTERNAL_ERROR | Server error |

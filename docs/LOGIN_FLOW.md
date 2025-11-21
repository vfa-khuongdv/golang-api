# Login Flow Documentation

## Simple Login Flow (Without MFA)

```
Client                           Server
  │                               │
  ├─ POST /api/v1/login ────────> │
  │  { email, password }          │
  │                               │
  │                               ├─ Find user by email
  │                               │
  │                               ├─ Verify password
  │                               │
  │                               ├─ Check if MFA enabled
  │                               │
  │                               ├─ Generate JWT tokens
  │                               │
  │ <─ 200 OK ──────────────────  │
  │  { access_token, refresh_token}
  │                               │
  ├─ Store tokens in storage      │
  │                               │
```

## Login Flow With MFA Enabled

```
Client                           Server
  │                               │
  ├─ POST /api/v1/login ────────> │
  │  { email, password }          │
  │                               │
  │                               ├─ Find user by email
  │                               │
  │                               ├─ Verify password
  │                               │
  │                               ├─ Check if MFA enabled
  │                               │
  │ <─ 200 OK ──────────────────  │
  │  { mfa_required: true,        │
  │    user_id: 1 }               │
  │                               │
  ├─ User enters TOTP code        │
  │                               │
  ├─ POST /api/v1/mfa/verify-code ├──> │
  │  { code: "123456",            │
  │    user_id: 1 }               │
  │                               │
  │                               ├─ Verify TOTP code
  │                               │
  │                               ├─ Generate JWT tokens
  │                               │
  │ <─ 200 OK ──────────────────  │
  │  { access_token, refresh_token}
  │                               │
  ├─ Store tokens in storage      │
  │                               │
```

## API Endpoints

- `POST /api/v1/login` - Login with email and password
- `POST /api/v1/refresh-token` - Refresh access token
- `POST /api/v1/mfa/verify-code` - Verify MFA code (if MFA enabled)
- `POST /api/v1/mfa/setup` - Initialize MFA setup
- `POST /api/v1/mfa/verify-setup` - Verify MFA setup with TOTP code
- `POST /api/v1/mfa/disable` - Disable MFA
- `GET /api/v1/mfa/status` - Check MFA status

## Login Request

```json
POST /api/v1/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

## Login Response (Without MFA)

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "mfa_required": false
}
```

## Login Response (With MFA Enabled)

```json
{
  "mfa_required": true,
  "user_id": 1,
  "message": "MFA code required"
}
```

## MFA Verification Request

```json
POST /api/v1/mfa/verify-code
Content-Type: application/json

{
  "code": "123456",
  "user_id": 1
}
```

## MFA Verification Response

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

## Refresh Token Request

```json
POST /api/v1/refresh-token
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Refresh Token Response

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

## MFA Setup Flow

### 1. Initialize MFA Setup

```json
POST /api/v1/mfa/setup
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "secret": "JBSWY3DPEBLW64TMMQ======"
}
```

### 2. Verify MFA Setup

```json
POST /api/v1/mfa/verify-setup
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "code": "123456"
}
```

**Response:**
```json
{
  "message": "MFA setup verified successfully"
}
```

## Using Access Token

Add the access token to the Authorization header:

```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer <access_token>"
```

## Error Codes

| Status | Error | Description |
|--------|-------|-------------|
| 400 | validation_error | Invalid input or validation failed |
| 401 | invalid_credentials | Invalid email or password |
| 401 | invalid_token | Invalid or expired token |
| 401 | invalid_mfa_code | Invalid MFA code |
| 404 | user_not_found | User doesn't exist |
| 500 | internal_error | Server error |

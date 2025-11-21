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
  │                               ├─ Generate JWT tokens (access scope)
  │                               │
  │ <─ 200 OK ──────────────────  │
  │  { access_token, refresh_token}
  │                               │
  ├─ Store tokens in storage      │
  │                               │
```

## Login Flow With MFA Enabled (Updated)

### Step 1: Initial Login - Get MFA Temporary Token

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
  │                               ├─ Generate temporary MFA token (10 min expiry)
  │                               │ (Scope: mfa_verification)
  │                               │
  │ <─ 200 OK ──────────────────  │
  │  { mfa_required: true,        │
  │    temporary_token: "...",    │
  │    user_id: 1 }               │
  │                               │
  ├─ User enters TOTP code        │
  │                               │
```

### Step 2: Verify MFA Code - Get Access Token

```
Client                           Server
  │                               │
  ├─ POST /api/v1/mfa/verify-code ──────> │
  │  Authorization: Bearer <temporary_token>
  │  { code: "123456" }           │
  │                               │
  │                               ├─ Validate MFA temporary token (scope check)
  │                               │
  │                               ├─ Verify TOTP code
  │                               │
  │                               ├─ Generate JWT tokens (access scope)
  │                               │ (Normal access token with 1 hour expiry)
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
  "access_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637433600
  },
  "refresh_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637520000
  }
}
```

## Login Response (With MFA Enabled)

```json
{
  "mfa_required": true,
  "temporary_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "MFA code required"
}
```

**Note:** The `temporary_token` has:
- Scope: `mfa_verification` (only valid for `/api/v1/mfa/verify-code` endpoint)
- TTL: 10 minutes
- Cannot be used to access other protected endpoints

## MFA Verification Request

```json
POST /api/v1/mfa/verify-code
Authorization: Bearer <temporary_token>
Content-Type: application/json

{
  "code": "123456"
}
```

**Requirements:**
- `Authorization` header with `temporary_token` (from login response)
- Token must have `mfa_verification` scope
- Token must not be expired (10-minute TTL)

## MFA Verification Response

```json
{
  "access_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637433600
  },
  "refresh_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637520000
  }
}
```

## Refresh Token Request

```json
POST /api/v1/refresh-token
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Requirements:**
- `refresh_token`: The existing refresh token (required, used to generate new tokens and must be valid in database)
- `access_token`: The existing access token (required, used to verify token ownership, can be expired)
- Both tokens must belong to the same user

**Security Notes:**
- Refresh token is validated against the database to prevent token reuse attacks
- Access token signature is verified even if expired, to ensure token ownership
- A mismatch between tokens (different users) will result in an error

## Refresh Token Response

```json
{
  "access_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637433600
  },
  "refresh_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1637520000
  }
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

## JWT Token Scopes

The system uses JWT token scopes to restrict token usage to specific endpoints:

### Access Token (Scope: `access`)
- **Purpose:** General API access for authenticated users
- **TTL:** 1 hour
- **Valid for:** All protected endpoints requiring authentication
  - `GET /api/v1/profile`
  - `POST /api/v1/mfa/setup`
  - `POST /api/v1/mfa/verify-setup`
  - `POST /api/v1/mfa/disable`
  - `GET /api/v1/mfa/status`
  - `POST /api/v1/change-password`
  - And all other protected endpoints
- **Invalid for:** `/api/v1/mfa/verify-code` (requires mfa_verification scope)

### MFA Verification Token (Scope: `mfa_verification`)
- **Purpose:** Temporary token for MFA verification during login
- **TTL:** 10 minutes
- **Valid for:** Only `/api/v1/mfa/verify-code` endpoint
- **Invalid for:** All other protected endpoints
- **When obtained:** After successful login with MFA enabled (step 1)
- **When consumed:** After successful TOTP code verification (step 2)

**Security Note:** A token with `access` scope CANNOT be used on `/api/v1/mfa/verify-code` and vice versa.

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
| 401 | invalid_token | Invalid or expired token (but works if access_token can verify signature) |
| 401 | invalid_mfa_code | Invalid MFA code |
| 404 | user_not_found | User doesn't exist |
| 409 | token_mismatch | Refresh and access tokens belong to different users |
| 500 | internal_error | Server error |

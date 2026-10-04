# Golang Template REST API

A Go REST API for a Content Management System (CMS) with user authentication, JWT access tokens, and refresh tokens. The project implements clean architecture with clear separation of concerns and comprehensive test coverage. It uses Gin and GORM on MySQL, Docker Compose for local dependencies (MySQL, phpMyAdmin, Mailpit), and includes SQL migrations, seeding, and email support.

## Key Features

- **User Authentication**: JWT access tokens (1 hour) and rotating refresh tokens (30 days, stored as SHA-256 hashes)
- **Account Protection**: Account lockout after 5 failed logins (15 minutes) and a per-IP rate limit (10 requests/minute) on the public auth endpoints
- **Role-Based Access Control**: Roles grant permissions; a user can only grant or remove permissions they hold
- **Request Limits**: Request bodies are limited to 1 MB (`413` when larger)
- **Password Management**: bcrypt hashing, password complexity validation, change password, and password reset via email (only a hash of the reset token is stored)
- **Email Service**: SMTP integration (go-mail) for password reset emails, configured through the `settings` table
- **Application Settings**: Key/value `settings` table, with secret values encrypted using AES-256-GCM
- **API Documentation**: OpenAPI 3.0 specification with Swagger UI (served when `STAGE` is not `prod`)
- **Database Migrations**: SQL migrations applied with golang-migrate on startup when `RUN_MIGRATE=true`
- **Structured Logging**: JSON logs with request IDs, event names, and masking of sensitive data (see [docs/logging-standards.md](docs/logging-standards.md))
- **Clean Architecture**: Clear separation of concerns with handlers, services, repositories, and models layers
- **Testing**: Unit tests colocated with the code (Testify, mocks, in-memory SQLite) plus end-to-end tests
- **Docker Support**: Docker Compose for MySQL, phpMyAdmin, and Mailpit, and a Dockerfile for the application
- **Live Reloading**: Air integration for development with hot reload capability

## Architecture

The project follows clean architecture principles with the following layers:

- **Handlers (HTTP)**: Parse requests, validate input format, call services, return HTTP responses
- **Services (Business Logic)**: Implement business rules, validation, error handling, and orchestrate repositories
- **Repositories (Data Access)**: Handle all database operations using GORM, return domain entities
- **Models (Domain)**: Define domain objects with GORM and JSON serialization tags
- **Middlewares**: Handle cross-cutting concerns like request IDs, authentication, CORS, request logging, rate limiting, and panic recovery

For detailed information on development guidelines and patterns, see [DEVELOPMENT.md](DEVELOPMENT.md).
For testing standards and best practices, see [TESTING.md](TESTING.md).

## Project Structure

The project follows a clean architecture and is organized into the following directories:

```
├── .air.toml                         # Air (live reload) configuration
├── .env.example                      # Environment variable template
├── .github                           # CI workflow, Copilot instructions, and architecture skill
├── AGENTS.md                         # Guidelines for AI coding agents
├── DEVELOPMENT.md                    # Development guidelines
├── Dockerfile                        # Docker configuration for the application
├── Makefile                          # Build, test, and development commands
├── README.md                         # Project documentation
├── TESTING.md                        # Testing standards
├── cmd                               # Command-line entry points
│   ├── seeder                        # Seeder for initial data population
│   │   └── seeder.go
│   └── server                        # Main entry point for the web server
│       └── main.go
├── docker-compose.yml                # Docker Compose configuration for MySQL, phpMyAdmin, and Mailpit
├── docs                              # API and logging documentation
│   ├── logging-standards.md          # Logging standards
│   ├── swagger.html                  # Swagger UI page
│   └── swagger.json                  # OpenAPI 3.0 specification
├── go.mod                            # Go module dependencies
├── go.sum                            # Go module checksums
├── internal                          # Core application logic
│   ├── configs                       # Environment configuration and database connection
│   ├── database
│   │   ├── migrations                # SQL migrations (golang-migrate, *.up.sql / *.down.sql)
│   │   └── seeders                   # Seed data (users)
│   ├── handlers                      # HTTP request handlers (auth, user, health)
│   ├── middlewares                   # Auth, CORS, logging, rate limiting, request ID
│   ├── models                        # GORM models (User, RefreshToken, Setting)
│   ├── repositories                  # Repositories for database access
│   ├── routes                        # Router setup and dependency wiring
│   ├── services                      # Business logic (auth, JWT, refresh token, user, mail)
│   └── shared                        # Shared code used across layers
│       ├── constants                 # Application constants (e.g. settings keys)
│       ├── dto                       # Request/response data transfer objects
│       └── utils                     # Helpers (bcrypt, crypto, validation, responses, masking)
├── pkg                               # Reusable packages
│   ├── apperror                      # Custom application errors and error codes
│   ├── logger                        # Structured JSON logger (logrus)
│   ├── mailer                        # SMTP sender and embedded email templates
│   └── migrator                      # golang-migrate wrapper
└── tests                             # Shared test support
    ├── e2e                           # End-to-end tests (in-memory SQLite)
    └── mocks                         # Testify mocks for services and repositories
```

Unit tests live next to the code they test (`*_test.go`).

## Prerequisites

Before getting started, ensure that you have the following installed:

- [Go](https://golang.org/dl/) 1.27 or later (see `go.mod`)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Docker Compose](https://docs.docker.com/compose/)
- [Make](https://www.gnu.org/software/make/) (Usually pre-installed on macOS and Linux)
- [MySQL](https://dev.mysql.com/downloads/mysql/) (or use Docker MySQL)

## Setup Instructions

### 1. Install Required Tools

Before doing anything else, install all required development tools using the following command:

```bash
# Install all required tools
make install-tools
```

This will install:
- golangci-lint (for linting)
- Air (for live reloading)
- gotestsum (for running tests with better formatting)

It also runs `go mod tidy`. Migrations are applied by the server itself through the golang-migrate library, so the `migrate` CLI is only needed if you want to create new migration files.

### 2. Clone the repository and setup environment

```bash
git clone git@github.com:vfa-khuongdv/golang-api.git
cd golang-api
cp .env.example .env
```

Edit `.env` with your configuration values (database credentials, `JWT_KEY`, `SETTINGS_ENCRYPTION_KEY`, etc.). `docker-compose.yml` reads `DB_DATABASE`, `DB_USERNAME`, and `DB_PASSWORD` from it. Mail settings live in the `settings` table, see [Application Settings](#application-settings-settings-table).

### 3. Start the local services with Docker

Docker Compose runs the dependencies only; the Go server itself is started separately (see [Running the Server](#6-running-the-server)):

```bash
docker-compose up -d
```

This will:

- Start a MySQL 8.0 container on port 3306 (data is stored in `mysql/db/data`).
- Start a phpMyAdmin container on port 8080 for database management.
- Start a Mailpit container that catches outgoing mail (SMTP on 1026, web UI on 8026).

To build and run the application as a container, use the provided `Dockerfile` (`docker build -t golang-cms .`) and pass the environment variables from `.env`.

### 4. Database Migrations

Schema changes are plain SQL files in `internal/database/migrations`, applied with [golang-migrate](https://github.com/golang-migrate/migrate). The server applies pending migrations on startup when `RUN_MIGRATE=true` (the path is relative, so start the server from the repository root). The current migrations create the `users`, `refresh_tokens`, and `settings` tables and seed the default settings rows.

To create a new migration file, install the [migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) and run:

```bash
migrate create -ext sql -dir internal/database/migrations -seq your_migration_name
```

For example, to create a feedback table migration:
```bash
migrate create -ext sql -dir internal/database/migrations -seq feedback_table
```

This will create two files:
- `XXXXXX_feedback_table.up.sql` (for applying the migration)
- `XXXXXX_feedback_table.down.sql` (for reverting the migration)

### 5. Seeding the Database

The seeder does not create tables, so start the server once with `RUN_MIGRATE=true` first. Then seed two sample users (`john@example.com`, an admin, and `jane@example.com`, both with password `password123`, or `SEED_USER_PASSWORD` when set):

```bash
go run cmd/seeder/seeder.go
```

These accounts are for local development only: with `STAGE=prod` the seeder refuses to run unless `SEED_USER_PASSWORD` is set. Running the seeder again logs an error for each user because the emails already exist.

### 6. Running the Server

The server will be available at `http://localhost:3000` by default.

**Option 1: Using Make (Recommended)**

```bash
make dev
```

This command will:
1. Install required tools (if not already installed)
2. Start the server with Air for live reloading (configured in `.air.toml`)

The server needs a running MySQL (see step 3) and a valid `.env`.

**Option 2: Using Air Directly**

[Air](https://github.com/air-verse/air) provides live-reloading capability which is great for development:

```bash
air
```

**Option 3: Direct Go Run**

If you prefer to run the server directly without live-reloading:

```bash
go run cmd/server/main.go
```

**Option 4: Build a binary**

```bash
make build        # outputs bin/server
./bin/server
```

### 7. Database Management - PHPMyAdmin

PHPMyAdmin is available for database management through a web interface:
- URL: `http://localhost:8080`
- Username: `root`
- Password: (use the `DB_PASSWORD` value from your `.env` file)

### 8. Local Mail - Mailpit

Mailpit catches every email the app sends locally, so nothing reaches a real inbox:

```bash
docker-compose up -d mailpit
```

The default rows seeded by the migrations already point at Mailpit (`127.0.0.1:1026`, no SMTP AUTH, STARTTLS only if offered), so no extra configuration is needed.

Open `http://localhost:8026` to read the captured emails (e.g. password reset links).

## Environment Variables

The application is configured through the environment variables below. `DB_USERNAME`, `DB_PASSWORD`, `DB_DATABASE`, `JWT_KEY`, and `SETTINGS_ENCRYPTION_KEY` are required; startup fails if any of them is missing. See `.env.example` for a complete template:

**App Configuration:**
- `APP_SERVICE` - Service name used in logs (default: golang-cms)
- `APP_VERSION` - Version reported by `/api/v1/version` and logs (default: dev)
- `RUN_MIGRATE` - Run database migrations on startup when `true` (default: false; `.env.example` sets it to `true`)

**Database Configuration:**
- `DB_HOST` - MySQL database host (default: 127.0.0.1)
- `DB_PORT` - MySQL port number (default: 3306)
- `DB_USERNAME` - MySQL database username (required)
- `DB_PASSWORD` - MySQL database password (required)
- `DB_DATABASE` - MySQL database name (required)
- `DB_MAX_OPEN_CONNS` - Maximum open connections in the pool, per task (default: 20)
- `DB_MAX_IDLE_CONNS` - Maximum idle connections in the pool, per task (default: 5)
- `DB_CONN_MAX_LIFETIME` - Maximum lifetime of a connection, e.g. `30m` (default: 30m)
- `DB_CONN_MAX_IDLE_TIME` - Maximum idle time of a connection, e.g. `5m` (default: 5m)

**Server Configuration:**
- `PORT` - Port number for the application server (default: 3000)
- `GIN_MODE` - Gin mode ("debug", "release", or "test", default: release)
- `STAGE` - Environment stage, e.g. "local", "dev", "prod" (default: dev). Also reported as `env` in logs. Swagger UI and `swagger.json` are not served when it is `prod`.
- `CORS_ALLOWED_ORIGINS` - Comma-separated allowed CORS origins, read on every request (default: http://localhost:5173)
- `TRUSTED_PROXIES` - Comma-separated CIDRs of trusted reverse proxies / load balancers (default: `0.0.0.0/0`, trust every peer). The client IP used by the rate limiter and stored on refresh tokens is read from `X-Forwarded-For` when the request comes from a trusted peer. Behind a load balancer (AWS ALB, nginx, Cloudflare) it must trust the proxy, otherwise all clients appear with the proxy's IP and the public auth endpoints allow only 10 requests/minute for all users together. The default works without knowing the proxy's IP, but a client can spoof `X-Forwarded-For` and dodge the per-IP rate limit; set the proxy CIDR (e.g. the VPC CIDR `10.0.0.0/16`) to prevent that. With `STAGE=prod` and `TRUSTED_PROXIES` unset, the server logs a warning at startup. Set it to an empty value when the app is exposed directly with no proxy.

**JWT Configuration:**
- `JWT_KEY` - Secret key for JWT token signing, at least 32 characters (required; the router refuses to start with a shorter key)

**Settings Encryption:**
- `SETTINGS_ENCRYPTION_KEY` - Key used to encrypt secret rows of the `settings` table such as `mail.password`, at least 32 characters (required). Changing it makes existing encrypted values unreadable, so re-encrypt them afterwards. Generate one with `openssl rand -base64 48`.

These can be set in the `.env` file or passed as environment variables. A sample `.env.example` file is provided in the repository.

### Application Settings (`settings` table)

Mail and frontend settings are stored as key/value rows in the `settings` table instead of environment variables. Migrations create the rows with defaults for local development with [Mailpit](#8-local-mail---mailpit); update them directly in the database for other environments:

| Key             | Description                                        | Default                 |
|-----------------|----------------------------------------------------|-------------------------|
| `mail.host`     | SMTP server host                                   | `127.0.0.1`             |
| `mail.port`     | SMTP server port                                   | `1026`                  |
| `mail.username` | SMTP username (empty = no SMTP AUTH, e.g. Mailpit) | (empty)                 |
| `mail.password` | SMTP password, **stored encrypted** (see below)    | (empty)                 |
| `mail.from`     | Email address used as sender                       | `noreply@example.com`   |
| `app.frontend_url`  | Frontend base URL used in password reset links     | `http://localhost:5173` |

```sql
UPDATE settings SET value = 'smtp.gmail.com' WHERE `key` = 'mail.host';
UPDATE settings SET value = '587' WHERE `key` = 'mail.port';
UPDATE settings SET value = 'user@example.com' WHERE `key` = 'mail.username';
```

`mail.password` is stored encrypted (AES-256-GCM with `SETTINGS_ENCRYPTION_KEY`); a plaintext value in the table is rejected when sending mail, so do not set it with SQL. Set it through the authenticated settings API, which encrypts it before saving. Leave it empty when the SMTP server needs no password:

```bash
curl -X PUT http://localhost:3000/api/v1/settings \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"mail_password": "your-smtp-password"}'
```

## API Documentation

The API is documented using OpenAPI 3.0 specification. You can access the documentation through:

- **Swagger UI**: `http://localhost:3000/swagger` or `http://localhost:3000/api-docs`
- **OpenAPI JSON**: `http://localhost:3000/docs/swagger.json`

These routes are only registered when `STAGE` is not `prod`, and the server must be started from the repository root so it can find the `docs/` files.

### Main API Endpoints

The server runs on port `3000` by default. All authenticated endpoints require a valid JWT access token in the `Authorization` header: `Bearer <token>`

Every response carries an `X-Request-ID` header (the client's value is reused if it is at most 128 letters, digits or `-_.:=`, otherwise a new UUID is generated). Errors are returned as `{"code": <int>, "message": "..."}`; validation errors also include a `fields` array. Error codes are defined in `pkg/apperror/codes.go`.

#### Health and Version (Public)
- `GET /healthz` - Health status check
- `GET /readyz` - Readiness check (pings the database, 503 when it is unreachable)
- `GET /api/v1/version` - API version, build time, and uptime

#### Authentication (Public)

These four endpoints share a per-IP rate limit of 10 requests per minute (`429` with `X-RateLimit-*` headers when exceeded). After 5 failed logins an account is locked for 15 minutes; once the lock ends, the count starts again from zero. Every failed login (unknown email, wrong password, locked account, even with the right password) gets the same `400` answer, so it does not reveal which emails are registered.

- `POST /api/v1/login` - User login (returns access and refresh tokens)
- `POST /api/v1/refresh-token` - Exchange a refresh token and the (possibly expired) access token for a new pair; the refresh token is rotated
- `POST /api/v1/forgot-password` - Request password reset email (always responds at once with the same message, whether or not the email exists; the email is sent in the background)
- `POST /api/v1/reset-password` - Reset password using the emailed token (valid for 1 hour); signs out every session

#### Session (Authenticated)
- `POST /api/v1/logout` - Revoke all refresh tokens of the authenticated user

#### User Profile (Authenticated)
- `GET /api/v1/profile` - Get authenticated user's profile
- `PATCH /api/v1/profile` - Update authenticated user's profile
- `POST /api/v1/change-password` - Change authenticated user's password; signs out every session

#### Settings (Permission Required)
- `GET /api/v1/settings` - Get mail and frontend settings (the mail password is never returned); needs `settings:read`
- `PUT /api/v1/settings` - Update the provided settings (the mail password is stored encrypted); needs `settings:update`

#### Roles and Permissions (Permission Required)
- `GET /api/v1/roles`, `GET /api/v1/roles/:id` - List roles or get one, with their permissions; needs `roles:read`
- `POST /api/v1/roles` - Create a role; needs `roles:create`
- `PUT /api/v1/roles/:id` - Update a role and replace its permissions; needs `roles:update`
- `DELETE /api/v1/roles/:id` - Delete a role; needs `roles:delete`
- `GET /api/v1/permissions` - List all permissions; needs `permissions:read`
- `PUT /api/v1/users/:id/roles` - Replace the roles of a user; needs `users:assign-roles`

Without the permission the answer is `403`. A user can only create, change, delete, assign or remove a role whose permissions they all hold, so nobody can grant more than they have. The built-in `admin` role holds every permission and cannot be changed or deleted, and the last admin cannot lose it.

## Testing

To install required testing tools and run tests with coverage report generation:

```bash
make test-coverage
```

This command will:
1. Install required tools (golangci-lint, air, gotestsum) if not already installed
2. Run the tests of the core packages (shared, handlers, middlewares, repositories, services, `pkg`) and generate `coverage.out`
3. Generate a coverage summary at `coverage-summary.txt`
4. Generate an HTML coverage report at `coverage.html`

For specific tests, you can still use:

```bash
go test -v path/to/test
```

### Other Testing Commands

- `make test`: Run all unit tests using gotestsum (excludes `cmd`, `docs`, and `tests`)
- `make test-e2e`: Run end-to-end tests (`tests/e2e`)
- `make watch-test`: Watch for changes and run tests automatically (requires [reflex](https://github.com/cespare/reflex))

### Test Layout

Unit tests sit next to the code they test (`internal/**/*_test.go`, `pkg/**/*_test.go`). `tests/e2e` holds end-to-end tests that run the real router against in-memory SQLite, and `tests/mocks` holds the shared Testify mocks. See [TESTING.md](TESTING.md) for details.

### Development Commands

- `make install-tools`: Install all required development tools
- `make help`: List all targets
- `make build`: Build the application binary to `bin/server`
- `make clean`: Remove generated files and binaries
- `make test`: Run unit tests with gotestsum
- `make test-e2e`: Run end-to-end tests
- `make test-coverage`: Run tests with coverage report generation (HTML and summary)
- `make watch-test`: Watch for changes and run tests automatically (requires reflex)
- `make lint`: Run linter (golangci-lint)
- `make fmt`: Format code using go fmt
- `make vet`: Run go vet static analysis
- `make pre-push`: Run all checks (fmt, vet, lint, test) before pushing
- `make dev`: Start server with Air (requires MySQL running)

## Contribution Guidelines

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/feature-name`).
3. Commit your changes (`git commit -am 'Add feature'`).
4. Push to the branch (`git push origin feature/feature-name`).
5. Open a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE-MIT](LICENSE-MIT) file for details.

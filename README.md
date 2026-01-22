# Golang Template REST API

A production-ready Go REST API for a Content Management System (CMS) with user authentication, JWT tokens, and refresh tokens. The project implements clean architecture with clear separation of concerns and comprehensive test coverage. It uses MySQL for persistence, Docker for containerization, and includes database migrations, seeding, and email support.

## Key Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **Password Management**: Secure password hashing with bcrypt, password reset via email
- **Email Service**: SMTP integration for sending password reset and other notification emails
- **API Documentation**: OpenAPI 3.0 specification with Swagger UI
- **Database Migrations**: Automated schema management with migration support
- **Clean Architecture**: Clear separation of concerns with handlers, services, repositories, and models layers
- **Comprehensive Testing**: Unit tests, integration tests, and end-to-end tests with high coverage
- **Docker Support**: Containerized application and MySQL database with Docker Compose
- **Live Reloading**: Air integration for development with hot reload capability

## Architecture

The project follows clean architecture principles with the following layers:

- **Handlers (HTTP)**: Parse requests, validate input format, call services, return HTTP responses
- **Services (Business Logic)**: Implement business rules, validation, error handling, and orchestrate repositories
- **Repositories (Data Access)**: Handle all database operations using GORM, return domain entities
- **Models (Domain)**: Define domain objects with GORM and JSON serialization tags
- **Middlewares**: Handle cross-cutting concerns like authentication, CORS, logging, and error recovery

For detailed information on development guidelines and patterns, see [DEVELOPMENT.md](DEVELOPMENT.md).
For testing standards and best practices, see [TESTING.md](TESTING.md).

## Project Structure

The project follows a clean architecture and is organized into the following directories:

```
├── Dockerfile                        # Docker configuration for the application
├── README.md                         # Project documentation
├── cmd                               # Command-line interfaces (CLI)
│   ├── seeder                        # Seeder for initial data population
│   │   └── seeder.go
│   └── server                        # Main entry point for the web server
│       └── main.go
├── docker-compose.yml                # Docker Compose configuration for the app and MySQL
├── docs                              # API documentation
│   ├── swagger.json                  # OpenAPI 3.0 specification
│   ├── swagger.html                  # Swagger UI documentation
│   └── LOGIN_FLOW.md                 # Login flow documentation
├── go.mod                            # Go module dependencies
├── go.sum                            # Go module checksums
├── internal                          # Core application logic
│   ├── configs                       # Configuration files for database, environment variables, JWT, etc.
│   ├── constants                     # Constants and error handling
│   ├── database                      # Database migrations and seeding
│   ├── dto                           # Data transfer objects for request and response
│   ├── handlers                      # HTTP request handlers
│   ├── middlewares                   # Middlewares for authentication and logging
│   ├── models                        # Data models for the application
│   ├── repositories                  # Repositories for database access
│   ├── routes                        # Routes and routing logic
│   ├── services                      # Business logic for authentication, user, etc.
│   └── utils                         # Utility functions (e.g., for encryption, validation)
├── pkg                               # External packages
│   ├── apperror                      # Custom application errors
│   ├── logger                        # Logger utility
│   ├── mailer                        # Mailer for sending emails
│   └── migrator                      # Database migration utility
├── tests                             # Unit and integration tests
│   ├── e2e                           # End-to-end tests
│   └── mocks                         # Mocks for internal package tests
```

## Prerequisites

Before getting started, ensure that you have the following installed:

- [Go](https://golang.org/dl/) (Go 1.21 or later; project targets Go 1.25.2)
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
- Migrate CLI (for database migrations)
- Air (for live reloading)
- gotestsum (for running tests with better formatting)

### 2. Clone the repository and setup environment

```bash
git clone git@github.com:vfa-khuongdv/golang-api.git
cd golang-api
cp .env.example .env
```

Edit `.env` with your configuration values (database credentials, JWT secret, SMTP settings, etc.).

### 3. Build and run the application using Docker

You can use Docker Compose to set up both the app and the MySQL database:

```bash
docker-compose up --build
```

This will:

- Build the Docker images.
- Start a MySQL container on port 3306.
- Start the application container on port 3000.
- Start a PHPMyAdmin container on port 8080 for database management.

### 4. Database Migrations

To create a new migration file, use the following command:

```bash
migrate create -ext sql -dir internal/database/migrations -seq your_migration_name
```

For example, to create a feedback table migration:
```bash
migrate create -ext sql -dir internal/database/migrations -seq feedback_table
```

This will create two files:
- XXXXXX_feedback_table.up.sql (for applying the migration)
- XXXXXX_feedback_table.down.sql (for reverting the migration)

The project includes migrations for creating the necessary tables in the MySQL database.
To apply the migrations:

```bash
make migrate
```

Or manually:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" up
```

To revert migrations, you can use the down command:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" down
```

You can also revert a specific number of migrations by adding the number after the down command:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" down 1
```

### 5. Seeding the Database

To seed the database with initial data (e.g., default users, roles, permissions), run:

```bash
make start-seeder
```

### 6. Running the Server

The server will be available at `http://localhost:3000` by default.

**Option 1: Using Make (Recommended)**

```bash
make start-server
```

This command will:
1. Install required tools (if not already installed)
2. Start Docker containers in detached mode
3. Start the server with Air for live reloading

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

**Option 4: Using Docker**

```bash
make docker-up
```

### 7. Database Management - PHPMyAdmin

PHPMyAdmin is available for database management through a web interface:
- URL: `http://localhost:8080`
- Username: `root`
- Password: (use the `DB_PASSWORD` value from your `.env` file)

## Environment Variables

The following environment variables are required for the application. See `.env.example` for a complete template:

**Database Configuration:**
- `DB_HOST` - MySQL database host (default: 127.0.0.1)
- `DB_PORT` - MySQL port number (default: 3306)
- `DB_USERNAME` - MySQL database username (default: user)
- `DB_PASSWORD` - MySQL database password (default: password)
- `DB_DATABASE` - MySQL database name (default: dbname)

**Server Configuration:**
- `APP_PORT` - Port number for the application server (default: 3000)
- `GIN_MODE` - Gin mode ("debug" or "release", default: debug)
- `STAGE` - Environment stage ("local", "dev", "prod", default: dev)

**JWT Configuration:**
- `JWT_SECRET` - Secret key for JWT token signing (required)
- `JWT_EXPIRY` - JWT token expiration in seconds (default: 900 / 15 minutes)
- `REFRESH_TOKEN_EXPIRY` - Refresh token expiration in seconds (default: 604800 / 7 days)

**SMTP/Email Configuration:**
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USER` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `MAIL_FROM` - Email address used as sender

**Frontend Configuration:**
- `FRONTEND_URL` - URL of the frontend application for password reset links

These can be set in the `.env` file or passed as environment variables. A sample `.env.example` file is provided in the repository.

## API Documentation

The API is documented using OpenAPI 3.0 specification. You can access the documentation through:

- **Swagger UI**: `http://localhost:3000/swagger` or `http://localhost:3000/api-docs`
- **OpenAPI JSON**: `http://localhost:3000/docs/swagger.json`

### Main API Endpoints

The server runs on port `3000` by default. All authenticated endpoints require a valid JWT token in the `Authorization` header: `Bearer <token>`

#### Health Check (Public)
- `GET /healthz` - Health status check

#### Authentication (Public)
- `POST /api/v1/login` - User login (returns access and refresh tokens)
- `POST /api/v1/refresh-token` - Refresh access token using refresh token
- `POST /api/v1/forgot-password` - Request password reset email
- `POST /api/v1/reset-password` - Reset password using reset token

#### User Profile (Authenticated)
- `GET /api/v1/profile` - Get authenticated user's profile
- `PATCH /api/v1/profile` - Update authenticated user's profile
- `POST /api/v1/change-password` - Change authenticated user's password

## Testing

To install required testing tools and run tests with coverage report generation:

```bash
make test-coverage
```

This command will:
1. Install required tools (gotestsum, gocov, gocov-html) if not already installed
2. Run all tests and generate coverage.out
3. Generate a coverage summary at coverage-summary.txt
4. Generate an HTML coverage report at coverage.html

For specific tests, you can still use:

```bash
go test -v path/to/test
```

### Other Testing Commands

- `make test`: Run all unit tests using gotestsum
- `make test-e2e`: Run end-to-end tests
- `make watch-test`: Watch for changes and run tests automatically

### Unit Tests Directory

The test files are located under the `tests` directory. The tests follow the Go testing conventions.

### Development Commands

- `make install-tools`: Install all required development tools
- `make build`: Build the application binary
- `make clean`: Remove generated files and binaries
- `make test`: Run unit tests with gotestsum
- `make test-e2e`: Run end-to-end tests
- `make test-coverage`: Run tests with coverage report generation (HTML and summary)
- `make watch-test`: Watch for changes and run tests automatically
- `make lint`: Run linter (golangci-lint)
- `make fmt`: Format code using go fmt
- `make vet`: Run go vet static analysis
- `make pre-push`: Run all checks (fmt, vet, lint, test) before pushing
- `make docker-up`: Start Docker containers
- `make wait-for-db`: Wait for MySQL to be ready
- `make dev`: Start server with Air (requires DB to be running)
- `make start-server`: Full dev environment setup (docker-up → wait-for-db → dev)
- `make start-seeder`: Seed the database with initial data
- `make migrate`: Run database migrations up
- `make migrate-down`: Revert database migrations down

## Contribution Guidelines

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/feature-name`).
3. Commit your changes (`git commit -am 'Add feature'`).
4. Push to the branch (`git push origin feature/feature-name`).
5. Open a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

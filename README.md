# Golang Template Project

This is a Golang project designed to handle a simple web service with user management, roles, permissions, and refresh tokens. It uses MySQL for the database, Docker for containerization, and includes support for migrations, seeding, and authentication.

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
│   ├── logger                        # Logger utility
│   └── mailer                        # Mailer for sending emails
├── tests                             # Unit and integration tests
│   ├── e2e                           # End-to-end tests
│   └── mocks                         # Mocks for internal package tests
```

## Prerequisites

Before getting started, ensure that you have the following installed:

- [Go](https://golang.org/dl/) (Go 1.23 or later)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Docker Compose](https://docs.docker.com/compose/)
- [MySQL](https://www.mysql.com/)
- [Makefile](https://www.gnu.org/software/make/) (Usually pre-installed on macOS and Linux)

## Setup Instructions

### 1. Install Required Tools

Before doing anything else, install all required development tools using the following command:

```bash
# Install all required tools
make install-tools
```

This will install:
- Migrate CLI (for database migrations)
- Air (for live reloading)
- gotestsum (for running tests)
- golangci-lint (for linting)

### 2. Clone the repository

```bash
git clone https://github.com/yourusername/yourproject.git
cd yourproject
cp .env.example .env
```

### 3. Build and run the application using Docker

You can use Docker Compose to set up both the app and the MySQL database:

```bash
docker-compose up --build
```

This will:

- Build the Docker images.
- Start a MySQL container.
- Start the application container.

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

The easiest way to run the server is using the provided make command:

```bash
make start-server
```

This command will:
1. Install required tools (if not already installed)
2. Start Docker containers in detached mode
3. Start the server with Air for live reloading

Alternatively, you can run the server in other ways:

#### Using Air Directly

[Air](https://github.com/air-verse/air) provides live-reloading capability which is great for development:

```bash
air
```

#### Direct Go Run

If you prefer to run the server directly without live-reloading:

```bash
go run cmd/server/main.go
```

The server will start and be available at `http://localhost:3000`.

### 7. phpmyadmin

PHPMyAdmin is available for database management through a web interface at:
- URL: `http://localhost:8080`
- Username: `root`
- Password: `root`

## Environment Variables

The following environment variables are required for the application:

Database Configuration:
- `DB_HOST` - MySQL database host
- `DB_PORT` - MySQL port number
- `DB_USERNAME` - MySQL database username
- `DB_PASSWORD` - MySQL database password
- `DB_DATABASE` - MySQL database name

Server Configuration:
- `PORT` - Port number for the application server (default: 3000)
- `GIN_MODE` - Gin mode ("debug" or "release")
- `RUN_MIGRATE` - Whether to run migrations on startup
- `STAGE` - Environment stage ("local", "dev", "prod")

JWT Configuration:
- `JWT_KEY` - Secret key for JWT token generation

URL Configuration:
- `FRONTEND_URL` - URL of the frontend application

Mail Configuration:
- `MAIL_HOST` - SMTP server host
- `MAIL_PORT` - SMTP server port
- `MAIL_USERNAME` - SMTP username
- `MAIL_PASSWORD` - SMTP password
- `MAIL_FROM` - Email address used as sender

These can be set in the `.env` file or passed directly as environment variables. A sample `.env.example` file is provided in the repository.

## API Documentation

The API is documented using OpenAPI 3.0 specification. You can access the documentation through:

- **Swagger UI**: `http://localhost:8080/swagger` or `http://localhost:8080/api-docs`
- **OpenAPI JSON**: `http://localhost:8080/docs/swagger.json`
- **Login Flow**: See `docs/LOGIN_FLOW.md` for detailed authentication flow

### Main API Endpoints

#### Authentication (Public)
- `POST /api/v1/login` - User login
- `POST /api/v1/refresh-token` - Refresh access token
- `POST /api/v1/forgot-password` - Request password reset
- `POST /api/v1/reset-password` - Reset password with token
- `POST /api/v1/mfa/verify-code` - Verify MFA code during login

#### User Profile (Authenticated)
- `GET /api/v1/profile` - Get user profile
- `PATCH /api/v1/profile` - Update user profile
- `POST /api/v1/change-password` - Change user password

#### User Management (Authenticated)
- `GET /api/v1/users` - List users (admin only)
- `POST /api/v1/users` - Create user (admin only)
- `GET /api/v1/users/{id}` - Get user by ID
- `PATCH /api/v1/users/{id}` - Update user (admin only)
- `DELETE /api/v1/users/{id}` - Delete user (admin only)

#### Multi-Factor Authentication (Authenticated)
- `POST /api/v1/mfa/setup` - Initialize MFA setup
- `POST /api/v1/mfa/verify-setup` - Verify MFA setup
- `POST /api/v1/mfa/disable` - Disable MFA
- `GET /api/v1/mfa/status` - Get MFA status

#### Health Check
- `GET /healthz` - Health status

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

## Contribution Guidelines

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/feature-name`).
3. Commit your changes (`git commit -am 'Add feature'`).
4. Push to the branch (`git push origin feature/feature-name`).
5. Open a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

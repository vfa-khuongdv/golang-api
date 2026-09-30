# AGENTS.md - Agent Coding Guidelines

> For detailed architecture, testing patterns, code templates, and conventions, see:
> - `.github/skills/golang-cms-architecture/SKILL.md` - Full development guidelines
> - `.github/skills/golang-cms-architecture/references/` - Templates, cheatsheet, commands
> - `README.md`, `DEVELOPMENT.md`, `TESTING.md` - Setup, coding standards, testing standards
> - `docs/logging-standards.md` - Logging fields, events, and data masking

## Project Overview

Go 1.27+ CMS REST API with JWT auth (access + refresh tokens), password reset by email, and clean architecture.
- **Framework:** Gin + GORM
- **Database:** MySQL 8.0+ (SQLite in-memory for tests)
- **Migrations:** SQL files in `internal/database/migrations` (golang-migrate)
- **Testing:** Testify with mocks

## Running the Application

```bash
# Configure the environment (DB_*, JWT_KEY and SETTINGS_ENCRYPTION_KEY are required)
cp .env.example .env

# Start MySQL and Mailpit (via Docker or native)
docker-compose up -d mysql mailpit

# Start server from the repository root (RUN_MIGRATE=true applies the migrations)
go run cmd/server/main.go

# Seed sample users (run after the tables exist)
go run cmd/seeder/seeder.go
```

> See `.github/skills/golang-cms-architecture/references/commands.md` for full command reference.

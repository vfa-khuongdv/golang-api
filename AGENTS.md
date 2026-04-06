# AGENTS.md - Agent Coding Guidelines

> For detailed architecture, testing patterns, code templates, and conventions, see:
> - `.github/skills/golang-cms-architecture/SKILL.md` - Full development guidelines
> - `.github/skills/golang-cms-architecture/references/` - Templates, cheatsheet, commands

## Project Overview

Go 1.25+ CMS REST API with JWT auth, MFA (TOTP), and clean architecture.
- **Framework:** Gin + GORM
- **Database:** MySQL 8.0+
- **Testing:** Testify with mocks

## Running the Application

```bash
# Start MySQL (via Docker or native)
docker-compose up -d mysql

# Run seeder
go run cmd/seeder/seeder.go

# Start server
go run cmd/server/main.go
```

> See `.github/skills/golang-cms-architecture/references/commands.md` for full command reference.

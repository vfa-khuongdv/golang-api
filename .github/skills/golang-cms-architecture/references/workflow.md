# Development Workflow

## Adding a New Feature

### 1. Model (internal/models/)
```go
type Feature struct {
    ID        uint      `gorm:"column:id;primaryKey" json:"id"`
    Name      string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
    IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
    CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Feature) TableName() string { return "features" }
```

Add a matching SQL migration (`*.up.sql` and `*.down.sql`) in `internal/database/migrations`; the schema is applied by golang-migrate, not by GORM AutoMigrate.

### 2. Repository (internal/repositories/)
- Define interface first
- Implement with GORM
- Add context.Context to all methods

### 3. Service (internal/services/)
- Business logic
- Use apperror for errors

### 4. DTOs (internal/shared/dto/)
- Request inputs with `binding` validation tags

### 5. Handler (internal/handlers/)
- Bind and validate the request (`utils.TranslateValidationErrors`)
- Call service
- Return response (`utils.RespondWithOK` / `utils.RespondWithError`)

### 6. Routes (internal/routes/routes.go)
- Create the repository, service, and handler in `SetupRouter` and register the route in the `public` or `authenticated` group

### 7. Mocks and Tests
- Add mocks for new interfaces in `tests/mocks`
- Unit tests for each layer, plus an e2e test in `tests/e2e`
- Target coverage: 85%+

### 8. Docs
- Update `docs/swagger.json` and the endpoint list in `README.md`

## Fixing a Bug

1. Reproduce with test
2. Identify layer (handler/service/repository)
3. Fix in appropriate layer
4. Verify with test

## Code Review Checklist

- [ ] Context passed through all layers?
- [ ] Errors handled with apperror?
- [ ] JSON tags in snake_case?
- [ ] Interfaces small (1-3 methods)?
- [ ] Migration added for schema changes?
- [ ] `docs/swagger.json` and README updated?
- [ ] Dependencies injected?
- [ ] Tests for success & failure cases?
- [ ] No sensitive data logged?
- [ ] No hardcoded values?

## Common Pitfalls

| Pitfall | Solution |
|---------|----------|
| Using `*gin.Context` in service | Use `context.Context` |
| Ignoring errors | Always handle with apperror |
| camelCase JSON tags | Use snake_case |
| Big interfaces | Split into smaller ones |
| No tests | Add tests immediately |
| Hardcoded config | Use env variables or the `settings` table |
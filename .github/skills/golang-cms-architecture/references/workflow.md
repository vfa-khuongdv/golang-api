# Development Workflow

## Adding a New Feature

### 1. Model (internal/models/)
```go
type Feature struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"size:255;not null"`
    IsActive  bool      `json:"is_active" gorm:"default:true"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 2. Repository (internal/repositories/)
- Define interface first
- Implement with GORM
- Add context.Context to all methods

### 3. Service (internal/services/)
- Business logic
- Validation
- Use apperror for errors

### 4. Handler (internal/handlers/)
- Parse request
- Call service
- Return response

### 5. Routes (internal/routes/)
- Register handler

### 6. Test
- Unit tests for each layer
- Target coverage: 85%+

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
| Hardcoded config | Use env variables |
# Code Templates

## New Repository

```go
package repositories

type ExampleRepository interface {
    Create(ctx context.Context, example *models.Example) (*models.Example, error)
    FindByField(ctx context.Context, field string, value string) (*models.Example, error)
    GetByID(ctx context.Context, id uint) (*models.Example, error)
    GetAll(ctx context.Context, page int, limit int) (*dto.Pagination[*models.Example], error)
    Update(ctx context.Context, example *models.Example) error
    Delete(ctx context.Context, id uint) error
}

type exampleRepositoryImpl struct {
    db *gorm.DB
}

func NewExampleRepository(db *gorm.DB) ExampleRepository {
    return &exampleRepositoryImpl{db: db}
}

func (repo *exampleRepositoryImpl) Create(ctx context.Context, example *models.Example) (*models.Example, error) {
    if err := repo.db.WithContext(ctx).Create(example).Error; err != nil {
        return nil, err
    }
    return example, nil
}

func (repo *exampleRepositoryImpl) GetByID(ctx context.Context, id uint) (*models.Example, error) {
    var example models.Example
    if err := repo.db.WithContext(ctx).First(&example, id).Error; err != nil {
        return nil, err
    }
    return &example, nil
}
```

## New Service

```go
package services

type ExampleService interface {
    CreateExample(ctx context.Context, input *dto.CreateExampleInput) (*models.Example, error)
    GetExample(ctx context.Context, id uint) (*models.Example, error)
}

type exampleServiceImpl struct {
    exampleRepo repositories.ExampleRepository
}

func NewExampleService(exampleRepo repositories.ExampleRepository) ExampleService {
    return &exampleServiceImpl{exampleRepo: exampleRepo}
}

func (svc *exampleServiceImpl) CreateExample(ctx context.Context, input *dto.CreateExampleInput) (*models.Example, error) {
    // Input format is already validated by the `binding` tags in the handler;
    // services enforce business rules.
    existing, err := svc.exampleRepo.FindByField(ctx, "name", input.Name)
    if err == nil && existing != nil {
        return nil, apperror.NewConflictError("Name already exists")
    }

    example := &models.Example{
        Name: input.Name,
    }

    created, err := svc.exampleRepo.Create(ctx, example)
    if err != nil {
        return nil, apperror.NewDBInsertError("Failed to create example")
    }
    return created, nil
}
```

## New Handler

```go
func (h *exampleHandler) CreateExample(c *gin.Context) {
    var input dto.CreateExampleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.RespondWithError(c, utils.TranslateValidationErrors(err, input))
        return
    }

    example, err := h.exampleService.CreateExample(c.Request.Context(), &input)
    if err != nil {
        utils.RespondWithError(c, err)
        return
    }

    utils.RespondWithOK(c, http.StatusCreated, example)
}
```

## DTO Templates

```go
type CreateExampleInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

// Pointer fields distinguish "not provided" from an empty value (see dto.UpdateProfileInput)
type UpdateExampleInput struct {
    Name  *string `json:"name" binding:"omitempty,min=1,max=255,not_blank"`
    Email *string `json:"email" binding:"omitempty,email"`
}

type ExampleResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Route Registration

Wire the new layers inside `SetupRouter` in `internal/routes/routes.go` and register the routes on the right group (`public` is rate limited; `authenticated` requires a JWT access token):

```go
// In SetupRouter, next to the existing repositories/services/handlers
exampleRepo := repositories.NewExampleRepository(db)
exampleService := services.NewExampleService(exampleRepo)
exampleHandler := handlers.NewExampleHandler(exampleService)

// Inside the `api` group
authenticated := api.Group("/")
authenticated.Use(middlewares.AuthMiddleware(jwtService))
{
    authenticated.POST("/examples", exampleHandler.CreateExample)
    authenticated.GET("/examples/:id", exampleHandler.GetExampleByID)
}
```

## Test Template (Service)

```go
func TestExampleService_CreateExample(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        mockRepo := new(mocks.MockExampleRepository)
        mockRepo.On("FindByField", mock.Anything, "name", "Test").Return((*models.Example)(nil), apperror.NewNotFoundError("not found"))
        mockRepo.On("Create", mock.Anything, mock.Anything).Return(&models.Example{ID: 1}, nil)
        
        svc := services.NewExampleService(mockRepo)
        result, err := svc.CreateExample(context.Background(), &dto.CreateExampleInput{Name: "Test"})
        
        require.NoError(t, err)
        assert.Equal(t, uint(1), result.ID)
        mockRepo.AssertExpectations(t)
    })
    
    t.Run("Conflict - Name Exists", func(t *testing.T) {
        mockRepo := new(mocks.MockExampleRepository)
        mockRepo.On("FindByField", mock.Anything, "name", "Test").Return(&models.Example{ID: 1}, nil)
        svc := services.NewExampleService(mockRepo)

        _, err := svc.CreateExample(context.Background(), &dto.CreateExampleInput{Name: "Test"})

        require.Error(t, err)
        appErr, ok := apperror.ToAppError(err)
        require.True(t, ok)
        assert.Equal(t, apperror.ErrConflict, appErr.Code)
        mockRepo.AssertExpectations(t)
    })
}
```

## Test Template (Handler)

```go
package handlers_test

func TestCreateExample(t *testing.T) {
    gin.SetMode(gin.TestMode)
    utils.InitValidator() // registers custom binding rules

    // post builds a test context with a JSON body
    post := func(body map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
        b, _ := json.Marshal(body)
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("POST", "/api/v1/examples", bytes.NewBuffer(b))
        c.Request.Header.Set("Content-Type", "application/json")
        c.Set("UserID", uint(1)) // only needed on authenticated routes
        return c, w
    }

    t.Run("Success", func(t *testing.T) {
        svc := new(mocks.MockExampleService)
        svc.On("CreateExample", mock.Anything, mock.AnythingOfType("*dto.CreateExampleInput")).
            Return(&models.Example{ID: 1, Name: "Test"}, nil)
        c, w := post(map[string]string{"name": "Test"})

        handlers.NewExampleHandler(svc).CreateExample(c)

        assert.Equal(t, http.StatusCreated, w.Code)
        svc.AssertExpectations(t)
    })

    t.Run("Validation Error", func(t *testing.T) {
        svc := new(mocks.MockExampleService) // service must not be called
        c, w := post(map[string]string{})

        handlers.NewExampleHandler(svc).CreateExample(c)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        svc.AssertExpectations(t)
    })

    t.Run("Service Error", func(t *testing.T) {
        svc := new(mocks.MockExampleService)
        svc.On("CreateExample", mock.Anything, mock.Anything).
            Return((*models.Example)(nil), apperror.NewInternalServerError("db error"))
        c, w := post(map[string]string{"name": "Test"})

        handlers.NewExampleHandler(svc).CreateExample(c)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
    })
}
```

For routes with path params set `c.Params = gin.Params{{Key: "id", Value: "1"}}`. Create the `recorder` first, then the context from it, and set up mocks before the request.

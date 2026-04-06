# Code Templates

## New Repository

```go
package repositories

type ExampleRepository interface {
    Create(ctx context.Context, example *models.Example) (*models.Example, error)
    GetByID(ctx context.Context, id uint) (*models.Example, error)
    GetAll(ctx context.Context, pagination *utils.Pagination) ([]models.Example, error)
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
    // Validate input
    if input.Name == "" {
        return nil, apperror.NewValidationError("Name is required")
    }
    
    example := &models.Example{
        Name: input.Name,
    }
    
    return svc.exampleRepo.Create(ctx, example)
}
```

## New Handler

```go
func (h *exampleHandler) CreateExample(c *gin.Context) {
    var input dto.CreateExampleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        utils.ResponseWithError(c, apperror.NewValidationError(err.Error()))
        return
    }

    example, err := h.exampleService.CreateExample(c.Request.Context(), &input)
    if err != nil {
        utils.ResponseWithError(c, err)
        return
    }

    utils.ResponseWithOK(c, http.StatusCreated, "Created successfully", example)
}
```

## DTO Templates

```go
type CreateExampleInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UpdateExampleInput struct {
    Name  string `json:"name"`
    Email string `json:"email" binding:"omitempty,email"`
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

```go
// In internal/routes/routes.go
func SetupRoutes(r *gin.Engine, handler *handlers.ExampleHandler) {
    api := r.Group("/api/v1")
    {
        examples := api.Group("/examples")
        {
            examples.POST("", handler.CreateExample)
            examples.GET("/:id", handler.GetExample)
            examples.PUT("/:id", handler.UpdateExample)
            examples.DELETE("/:id", handler.DeleteExample)
        }
    }
}
```

## Test Template (Service)

```go
func TestExampleService_CreateExample(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        mockRepo := new(mocks.MockExampleRepository)
        mockRepo.On("Create", mock.Anything, mock.Anything).Return(&models.Example{ID: 1}, nil)
        
        svc := services.NewExampleService(mockRepo)
        result, err := svc.CreateExample(context.Background(), &dto.CreateExampleInput{Name: "Test"})
        
        require.NoError(t, err)
        assert.Equal(t, uint(1), result.ID)
        mockRepo.AssertExpectations(t)
    })
    
    t.Run("Validation Error - Empty Name", func(t *testing.T) {
        mockRepo := new(mocks.MockExampleRepository)
        svc := services.NewExampleService(mockRepo)
        
        _, err := svc.CreateExample(context.Background(), &dto.CreateExampleInput{Name: ""})
        
        require.Error(t, err)
        assert.Contains(t, err.Error(), "Name is required")
    })
}
```
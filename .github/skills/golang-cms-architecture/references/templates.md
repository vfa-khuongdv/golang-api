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

## Test Template (Handler)

```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/vfa-khuongdv/golang-cms/internal/handlers"
    "github.com/vfa-khuongdv/golang-cms/internal/models"
    "github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
    "github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
    "github.com/vfa-khuongdv/golang-cms/pkg/apperror"
    "github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestCreateExample(t *testing.T) {
    gin.SetMode(gin.TestMode)
    utils.InitValidator()

    t.Run("CreateExample - Success", func(t *testing.T) {
        // ARRANGE - Setup mock and handler
        mockSvc := new(mocks.MockExampleService)
        handler := handlers.NewExampleHandler(mockSvc)

        example := &models.Example{ID: 1, Name: "Test"}
        mockSvc.On("CreateExample", mock.Anything, mock.AnythingOfType("*dto.CreateExampleInput")).Return(example, nil)

        // ARRANGE - Create request
        body, _ := json.Marshal(map[string]string{"name": "Test"})
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("POST", "/api/v1/examples", bytes.NewBuffer(body))
        c.Request.Header.Set("Content-Type", "application/json")

        // ACT - Call handler
        handler.CreateExample(c)

        // ASSERT - Verify response
        assert.Equal(t, http.StatusCreated, w.Code)
        mockSvc.AssertExpectations(t)
    })

    t.Run("CreateExample - Validation Error", func(t *testing.T) {
        mockSvc := new(mocks.MockExampleService)
        handler := handlers.NewExampleHandler(mockSvc)

        body, _ := json.Marshal(map[string]string{})
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("POST", "/api/v1/examples", bytes.NewBuffer(body))
        c.Request.Header.Set("Content-Type", "application/json")

        handler.CreateExample(c)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        mockSvc.AssertExpectations(t)
    })

    t.Run("CreateExample - Service Error", func(t *testing.T) {
        mockSvc := new(mocks.MockExampleService)
        handler := handlers.NewExampleHandler(mockSvc)

        mockSvc.On("CreateExample", mock.Anything, mock.AnythingOfType("*dto.CreateExampleInput")).Return(nil, apperror.NewInternalServerError("db error"))

        body, _ := json.Marshal(map[string]string{"name": "Test"})
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request, _ = http.NewRequest("POST", "/api/v1/examples", bytes.NewBuffer(body))
        c.Request.Header.Set("Content-Type", "application/json")

        handler.CreateExample(c)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        mockSvc.AssertExpectations(t)
    })
}

func TestGetExampleByID(t *testing.T) {
    gin.SetMode(gin.TestMode)
    utils.InitValidator()

    t.Run("GetExampleByID - Success", func(t *testing.T) {
        mockSvc := new(mocks.MockExampleService)
        handler := handlers.NewExampleHandler(mockSvc)

        example := &models.Example{ID: 1, Name: "Test"}
        mockSvc.On("GetExampleByID", mock.Anything, uint(1)).Return(example, nil)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Params = gin.Params{{Key: "id", Value: "1"}}
        c.Request, _ = http.NewRequest("GET", "/api/v1/examples/1", nil)

        handler.GetExampleByID(c)

        assert.Equal(t, http.StatusOK, w.Code)
        mockSvc.AssertExpectations(t)
    })

    t.Run("GetExampleByID - Invalid ID", func(t *testing.T) {
        mockSvc := new(mocks.MockExampleService)
        handler := handlers.NewExampleHandler(mockSvc)

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Params = gin.Params{{Key: "id", Value: "abc"}}
        c.Request, _ = http.NewRequest("GET", "/api/v1/examples/abc", nil)

        handler.GetExampleByID(c)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        mockSvc.AssertExpectations(t)
    })
}
```

### Handler Test Key Points

1. **Package**: Must be `handlers_test` (external test package)
2. **Context setup**: Always use `gin.SetMode(gin.TestMode)` and `utils.InitValidator()` at the start
3. **Request setup**: Create `httptest.NewRecorder()` FIRST, then `gin.CreateTestContext(w)` with that recorder
4. **Order**: ARRANGE (mock + request) → ACT (handler call) → ASSERT (verify)
5. **Mock location**: Setup mocks BEFORE creating the request/context
6. **Test structure**: One `t.Run` per test case, NOT nested test functions in wrong order
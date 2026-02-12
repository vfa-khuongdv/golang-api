package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestUpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize the validator
	utils.InitValidator()

	t.Run("UpdateProfile - Success", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		userID := uint(1)
		requestBody := map[string]any{
			"name":     "Updated User",
			"birthday": "2000-01-01",
			"address":  "456 New Street",
			"gender":   2,
		}
		input := &dto.UpdateProfileInput{
			Name:     utils.StringToPtr(requestBody["name"].(string)),
			Birthday: utils.StringToPtr(requestBody["birthday"].(string)),
			Address:  utils.StringToPtr(requestBody["address"].(string)),
			Gender:   utils.IntToPtr(int16(requestBody["gender"].(int))),
		}

		// Mock the service methods
		userService.On("UpdateProfile", userID, input).Return(nil)

		body, _ := json.Marshal(requestBody)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/profile", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the handler
		handler.UpdateProfile(c)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Update profile successfully"}`, w.Body.String())

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("UpdateProfile - Validation Error", func(t *testing.T) {
		tests := []struct {
			name           string
			reqBody        string
			expectedCode   float64
			expectedMsg    string
			expectedFields []apperror.FieldError
		}{
			{
				name:         "EmptyName",
				reqBody:      `{"name":""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "name", Message: "name must be at least 1 characters long or numeric"},
				},
			},
			{
				name:         "NameNotBlank",
				reqBody:      `{"name":"  "}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "name", Message: "name must not be blank"},
				},
			},
			{
				name:         "LongName",
				reqBody:      `{"name": "` + strings.Repeat("a", 46) + `"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "name", Message: "name must be at most 45 characters long or numeric"},
				},
			},
			{
				name:         "InvalidBirthdayFormat",
				reqBody:      `{"name": "User", "birthday": "invalid-date"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "birthday", Message: "birthday must be a valid date (YYYY-MM-DD) and not in the future"},
				},
			},
			{
				name:         "FutureBirthday",
				reqBody:      `{"name": "User", "birthday": "3000-01-01"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "birthday", Message: "birthday must be a valid date (YYYY-MM-DD) and not in the future"},
				},
			},
			{
				name:         "EmptyAddress",
				reqBody:      `{"name": "User", "birthday": "2000-01-01", "address": ""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "address", Message: "address must be at least 1 characters long or numeric"},
				},
			},
			{
				name:         "LongAddress",
				reqBody:      `{"name": "User", "birthday": "2000-01-01", "address": "` + strings.Repeat("a", 256) + `"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "address", Message: "address must be at most 255 characters long or numeric"},
				},
			},
			{
				name:         "AddressNotBlank",
				reqBody:      `{"name": "User", "birthday": "2000-01-01", "address": "  "}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "address", Message: "address must not be blank"},
				},
			},
			{
				name:         "InvalidGender 0",
				reqBody:      `{"name": "User", "birthday": "2000-01-01", "address": "123 Street", "gender": 0}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "gender", Message: "gender must be one of [1 2 3]"},
				},
			},
			{
				name:         "InvalidGender 4",
				reqBody:      `{"name": "User", "birthday": "2000-01-01", "address": "123 Street", "gender": 4}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "gender", Message: "gender must be one of [1 2 3]"},
				},
			},
			{
				name:           "StringGender",
				reqBody:        `{"name": "User", "birthday": "2000-01-01", "address": "123 Street", "gender": "male"}`,
				expectedCode:   float64(4001),
				expectedMsg:    "json: cannot unmarshal string into Go struct field UpdateProfileInput.gender of type int16",
				expectedFields: nil, // specific error case
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userService := new(mocks.MockUserService)
				mailerService := new(mocks.MockMailerService)
				handler := handlers.NewUserHandler(userService, mailerService)

				// Create a test context
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("PUT", "/api/v1/profile", bytes.NewBufferString(tt.reqBody))
				c.Set("UserID", uint(1))

				// Call the handler
				handler.UpdateProfile(c)

				// Assert the response
				expectedBody := map[string]any{
					"code":    tt.expectedCode,
					"message": tt.expectedMsg,
					"fields":  tt.expectedFields,
				}

				var actualBody map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &actualBody)

				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, expectedBody["code"], actualBody["code"])
				assert.Equal(t, expectedBody["message"], actualBody["message"])
				assert.Equal(t, tt.expectedFields, utils.ToFieldErrors(actualBody["fields"]))

				// Assert mocks
				userService.AssertExpectations(t)
				mailerService.AssertExpectations(t)
			})
		}
	})

	t.Run("UpdateProfile - Invalid UserID ctx", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/profile", nil)
		c.Set("UserID", 0) // Invalid User ID

		// Call the handler
		handler.UpdateProfile(c)

		// Assert the response
		var expectedBody = map[string]any{
			"code":    float64(apperror.ErrParseError),
			"message": "Invalid UserID",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("UpdateProfile - User Not Found", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		userID := uint(1)
		requestBody := map[string]any{
			"name":     "Updated User",
			"birthday": "2000-01-01",
			"address":  "456 New Street",
			"gender":   2,
		}
		input := &dto.UpdateProfileInput{
			Name:     utils.StringToPtr(requestBody["name"].(string)),
			Birthday: utils.StringToPtr(requestBody["birthday"].(string)),
			Address:  utils.StringToPtr(requestBody["address"].(string)),
			Gender:   utils.IntToPtr(int16(requestBody["gender"].(int))),
		}

		body, _ := json.Marshal(requestBody)

		// Mock the service method
		userService.On("UpdateProfile", userID, input).Return(apperror.NewNotFoundError("User not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/profile", bytes.NewBuffer(body))
		c.Set("UserID", userID)

		// Call the handler
		handler.UpdateProfile(c)

		// Assert the response
		var expectedBody = map[string]any{
			"code":    float64(apperror.ErrNotFound),
			"message": "User not found",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Error Update User", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		userID := uint(1)
		requestBody := map[string]any{
			"name":     "Updated User",
			"birthday": "2000-01-01",
			"address":  "456 New Street",
			"gender":   2,
		}
		input := &dto.UpdateProfileInput{
			Name:     utils.StringToPtr(requestBody["name"].(string)),
			Birthday: utils.StringToPtr(requestBody["birthday"].(string)),
			Address:  utils.StringToPtr(requestBody["address"].(string)),
			Gender:   utils.IntToPtr(int16(requestBody["gender"].(int))),
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service method
		userService.On("UpdateProfile", userID, input).Return(apperror.NewDBUpdateError("Update error"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/profile", bytes.NewBuffer(body))
		c.Set("UserID", userID)

		// Call the handler
		handler.UpdateProfile(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrDBUpdate),
			"message": "Update error",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

}

func TestGetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success get profile from database", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		user := &models.User{
			ID:        1,
			Email:     "email@example.com",
			Name:      "User",
			Gender:    1,
			CreatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		}
		// Mock the service method
		userService.On("GetProfile", uint(1)).Return(user, nil)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
		c.Set("UserID", uint(1))

		// Call the handler
		handler.GetProfile(c)

		// Assert the response
		expectedBody := map[string]any{
			"id":         float64(1),
			"email":      "email@example.com",
			"name":       "User",
			"gender":     float64(1),
			"created_at": "2023-10-01T00:00:00Z",
			"updated_at": "2023-10-01T00:00:00Z",
			"deleted_at": nil,
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedBody, actualBody)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Success get profile from redis cache", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		user := &models.User{
			ID:        1,
			Email:     "email@example.com",
			Name:      "User",
			Gender:    1,
			CreatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		}
		// Mock the service to return the cached profile
		userService.On("GetProfile", uint(1)).Return(user, nil)

		handler := handlers.NewUserHandler(userService, mailerService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
		c.Set("UserID", uint(1))

		// Call the handler
		handler.GetProfile(c)

		// Assert the response
		expectedBody := map[string]any{
			"id":         float64(1),
			"email":      "email@example.com",
			"name":       "User",
			"gender":     float64(1),
			"created_at": "2023-10-01T00:00:00Z",
			"updated_at": "2023-10-01T00:00:00Z",
			"deleted_at": nil,
		}

		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedBody, actualBody)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Error Invalid User ID", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)

		handler := handlers.NewUserHandler(userService, mailerService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
		c.Set("UserID", "invalid") // Invalid User ID

		// Call the GetProfile handler
		handler.GetProfile(c)

		var expectedBody = map[string]any{
			"code":    float64(apperror.ErrParseError),
			"message": "Invalid UserID",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Error User Not Found", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		userId := uint(1)

		userService.On("GetProfile", userId).Return(&models.User{}, apperror.NewNotFoundError("User not found"))

		handler := handlers.NewUserHandler(userService, mailerService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
		c.Set("UserID", userId)

		// Call the GetProfile handler
		handler.GetProfile(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrNotFound),
			"message": "User not found",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Success Get Profile but Error Cache", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)

		// Mock the GetUser method to return a user
		user := &models.User{
			ID:        1,
			Email:     "email@example.com",
			Name:      "User",
			Gender:    1,
			CreatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		}
		userService.On("GetProfile", uint(1)).Return(user, nil)

		handler := handlers.NewUserHandler(userService, mailerService)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)
		c.Set("UserID", uint(1))
		// Call the GetProfile handler
		handler.GetProfile(c)
		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		expectedBody := map[string]any{
			"id":         float64(1),
			"email":      "email@example.com",
			"name":       "User",
			"gender":     float64(1),
			"created_at": "2023-10-01T00:00:00Z",
			"updated_at": "2023-10-01T00:00:00Z",
			"deleted_at": nil,
		}

		var actualBody map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &actualBody)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedBody, actualBody)

	})
}

func TestChangePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize the validator
	utils.InitValidator()

	t.Run("ChangePassword - Success", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		user := &models.User{
			ID:        1,
			Email:     "email@example.com",
			Name:      "User",
			Password:  "$2a$10$I/L5VegpCyOlJPoa1.KrmeCdezSBIandsEL5S2dd4Ap0YIWk0Iuka", // bcrypt hash of "12345678"
			Gender:    1,
			CreatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		}
		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "newpassword",
			"confirm_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the services methods
		userService.On("ChangePassword", uint(1), mock.MatchedBy(func(input *dto.ChangePasswordInput) bool {
			return input.OldPassword == "12345678" &&
				input.NewPassword == "newpassword" &&
				input.ConfirmPassword == "newpassword"
		})).Return(user, nil)

		// Create http request and context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the ChangePassword handler
		handler.ChangePassword(c)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Change password successfully"}`, w.Body.String())

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - Validation Error", func(t *testing.T) {
		tests := []struct {
			name           string
			reqBody        string
			expectedCode   float64
			expectedMsg    string
			expectedFields []apperror.FieldError
		}{
			{
				name:         "EmptyOldPassword",
				reqBody:      `{"old_password":""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "old_password", Message: "old_password is required"},
					{Field: "new_password", Message: "new_password is required"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "ShortOldPassword",
				reqBody:      `{"old_password":"short"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "old_password", Message: "old_password must be at least 6 characters long or numeric"},
					{Field: "new_password", Message: "new_password is required"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "LongOldPassword",
				reqBody:      `{"old_password":"` + strings.Repeat("a", 256) + `"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "old_password", Message: "old_password must be at most 255 characters long or numeric"},
					{Field: "new_password", Message: "new_password is required"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "EmptyNewPassword",
				reqBody:      `{"old_password":"12345678","new_password":""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "new_password", Message: "new_password is required"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "ShortNewPassword",
				reqBody:      `{"old_password":"12345678","new_password":"short"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "new_password", Message: "new_password must be at least 6 characters long or numeric"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "LongNewPassword",
				reqBody:      `{"old_password":"12345678","new_password":"` + strings.Repeat("a", 256) + `"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "new_password", Message: "new_password must be at most 255 characters long or numeric"},
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "EmptyConfirmPassword",
				reqBody:      `{"old_password":"12345678","new_password":"newpassword","confirm_password":""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "confirm_password", Message: "confirm_password is required"},
				},
			},
			{
				name:         "ShortConfirmPassword",
				reqBody:      `{"old_password":"12345678","new_password":"newpassword","confirm_password":"short"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "confirm_password", Message: "confirm_password must be at least 6 characters long or numeric"},
				},
			},
			{
				name:         "LongConfirmPassword",
				reqBody:      `{"old_password":"12345678","new_password":"newpassword","confirm_password":"` + strings.Repeat("a", 256) + `"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "confirm_password", Message: "confirm_password must be at most 255 characters long or numeric"},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userService := new(mocks.MockUserService)
				mailerService := new(mocks.MockMailerService)
				handler := handlers.NewUserHandler(userService, mailerService)

				// Create http request and context
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBufferString(tt.reqBody))
				c.Set("UserID", uint(1))

				// Call the ChangePassword handler
				handler.ChangePassword(c)

				// Assert the response
				expectedBody := map[string]any{
					"code":    tt.expectedCode,
					"message": tt.expectedMsg,
					"fields":  tt.expectedFields,
				}
				var actualBody map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, expectedBody["code"], actualBody["code"])
				assert.Equal(t, expectedBody["message"], actualBody["message"])
				assert.Equal(t, expectedBody["fields"], utils.ToFieldErrors(actualBody["fields"]))

				// Assert mock expectations
				userService.AssertExpectations(t)
				mailerService.AssertExpectations(t)
			})
		}
	})

	t.Run("ChangePassword - NotFound User", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "newpassword",
			"confirm_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the ChangePassword method to return an error
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.NewNotFoundError("User not found"))

		// Create http request and context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the ChangePassword handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrNotFound),
			"message": "User not found",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mock expectations
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - Old Password Mismatch", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "wrongpassword",
			"new_password":     "newpassword",
			"confirm_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.NewInvalidPasswordError("Old password is incorrect"))

		// Create a new UserHandler instance
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the ChangePassword handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrInvalidPassword),
			"message": "Old password is incorrect",
		}

		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mock expectations
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - New Password and Confirm Password Mismatch", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "123456789",
			"confirm_password": "differentpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.NewPasswordMismatchError("New password and confirm password do not match"))

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the ChangePassword handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrPasswordMismatch),
			"message": "New password and confirm password do not match",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mock expectations
		userService.AssertExpectations(t)
	})

	t.Run("ChangePassword - Failed To Update", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "newpassword",
			"confirm_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.NewDBUpdateError("Update error"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrDBUpdate),
			"message": "Update error",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mock expectations
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - User Not found from ctx", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", nil)
		c.Set("UserID", "invalid") // Invalid User ID

		// Call the handler
		handler.ChangePassword(c)

		// Assert the response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, fmt.Sprintf(`{"code":%d,"message":"Invalid UserID"}`, apperror.ErrParseError), w.Body.String())

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - Old Password equal to New Password", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "12345678",
			"confirm_password": "12345678",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.NewPasswordMismatchError("New password must be different from old password"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrPasswordMismatch),
			"message": "New password must be different from old password",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ChangePassword - Hash Password Failed", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"old_password":     "12345678",
			"new_password":     "newpassword",
			"confirm_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ChangePassword", uint(1), mock.AnythingOfType("*dto.ChangePasswordInput")).Return(&models.User{}, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Hash password failed", nil))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/v1/change-password", bytes.NewBuffer(body))
		c.Set("UserID", uint(1))

		// Call the handler
		handler.ChangePassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrInternalServer),
			"message": "Hash password failed",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mock expectations
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

}

func TestResetPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	utils.InitValidator()

	t.Run("ResetPassword - Success", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"token":        "token",
			"new_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ResetPassword", mock.AnythingOfType("*dto.ResetPasswordInput")).Return(&models.User{}, nil)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(body))

		// Call the  handler
		handler.ResetPassword(c)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Reset password successfully"}`, w.Body.String())

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ResetPassword - Not found user by token", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"token":        "invalid-token",
			"password":     "newpassword",
			"new_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service method's behavior
		userService.On("ResetPassword", mock.AnythingOfType("*dto.ResetPasswordInput")).Return(&models.User{}, apperror.NewNotFoundError("User not found"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(body))

		// Call the handler
		handler.ResetPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrNotFound),
			"message": "User not found",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ResetPassword - Token Expired", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"token":        "invalid-token",
			"new_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service method
		userService.On("ResetPassword", mock.AnythingOfType("*dto.ResetPasswordInput")).Return(&models.User{}, apperror.Wrap(http.StatusBadRequest, apperror.ErrTokenExpired, "Token is expired", nil))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(body))

		// Call the handler
		handler.ResetPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrTokenExpired),
			"message": "Token is expired",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ResetPassword - Error Hashing Password Failed", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"token":        "token",
			"password":     "newpassword",
			"new_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ResetPassword", mock.AnythingOfType("*dto.ResetPasswordInput")).Return(&models.User{}, apperror.Wrap(http.StatusInternalServerError, apperror.ErrPasswordHashFailed, "Failed to hash password", nil))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(body))

		// Call the handler
		handler.ResetPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrPasswordHashFailed),
			"message": "Failed to hash password",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Error failed to UpdateUser", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"token":        "token",
			"password":     "newpassword",
			"new_password": "newpassword",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ResetPassword", mock.AnythingOfType("*dto.ResetPasswordInput")).Return(&models.User{}, apperror.NewDBUpdateError("Failed to update user"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("Authorization", "Bearer test-token")
		c.Set("UserID", uint(1))

		// Call the handler
		handler.ResetPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrDBUpdate),
			"message": "Failed to update user",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		tests := []struct {
			name          string
			reqBody       string
			expectedCode  float64
			expectedMsg   string
			expectedField []apperror.FieldError
		}{
			{
				name:         "EmptyToken",
				reqBody:      `{"token":""}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "token",
						Message: "token is required",
					},
					{
						Field:   "new_password",
						Message: "new_password is required",
					},
				},
			},
			{
				name:         "EmptyPassword",
				reqBody:      `{"token":"valid-token","password":""}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password is required",
					},
				},
			},
			{
				name:         "PasswordTooShort",
				reqBody:      `{"token":"valid-token","password":"short"}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password is required",
					},
				},
			},
			{
				name:         "PasswordTooLong",
				reqBody:      `{"token":"valid-token","password":"` + strings.Repeat("a", 256) + `"}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password is required",
					},
				},
			},
			{
				name:         "EmptyNewPassword",
				reqBody:      `{"token":"valid-token","password":"newpassword","new_password":""}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password is required",
					},
				},
			},
			{
				name:         "NewPasswordTooShort",
				reqBody:      `{"token":"valid-token","password":"newpassword","new_password":"short"}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password must be at least 6 characters long or numeric",
					},
				},
			},
			{
				name:         "NewPasswordTooLong",
				reqBody:      `{"token":"valid-token","password":"newpassword","new_password":"` + strings.Repeat("a", 256) + `"}`,
				expectedCode: 4001,
				expectedMsg:  "Validation failed",
				expectedField: []apperror.FieldError{
					{
						Field:   "new_password",
						Message: "new_password must be at most 255 characters long or numeric",
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userService := new(mocks.MockUserService)
				mailerService := new(mocks.MockMailerService)
				handler := handlers.NewUserHandler(userService, mailerService)

				// Create a test context
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBufferString(tt.reqBody))
				c.Set("UserID", uint(1))

				// Call the handler
				handler.ResetPassword(c)

				// Assert the response
				var actualBody map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, tt.expectedCode, actualBody["code"])
				assert.Equal(t, tt.expectedMsg, actualBody["message"])
				assert.Equal(t, tt.expectedField, utils.ToFieldErrors(actualBody["fields"]))

				// Assert mocks
				userService.AssertExpectations(t)
				mailerService.AssertExpectations(t)
			})
		}
	})

}

func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize the validator
	utils.InitValidator()

	t.Run("ForgotPassword - Success", func(t *testing.T) {
		_ = os.Setenv("MAIL_HOST", "smtp.gmail.com")
		_ = os.Setenv("MAIL_PORT", "587")
		_ = os.Setenv("MAIL_USERNAME", "test@example.com")
		_ = os.Setenv("MAIL_PASSWORD", "testpassword")
		_ = os.Setenv("MAIL_FROM", "noreply@example.com")
		_ = os.Setenv("FRONTEND_URL", "https://example.com")
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}

		requestBody := map[string]any{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ForgotPassword", mock.AnythingOfType("*dto.ForgotPasswordInput")).Return(user, nil)
		mailerService.On("SendMailForgotPassword", user).Return(nil)

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(body))

		handler.ForgotPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &responseBody)
		message := responseBody["message"].(string)

		assert.Equal(t, "If your email is in our system, you will receive instructions to reset your password", message)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ForgotPassword - Validation Error", func(t *testing.T) {
		tests := []struct {
			name           string
			reqBody        string
			expectedCode   float64
			expectedMsg    string
			expectedFields []apperror.FieldError
		}{
			{
				name:         "EmptyEmail",
				reqBody:      `{"email":""}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "email", Message: "email is required"},
				},
			},
			{
				name:         "InvalidEmailFormat",
				reqBody:      `{"email":"not-an-email"}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "email", Message: "email must be a valid email address"},
				},
			},
			{
				name:         "MissingEmail",
				reqBody:      `{}`,
				expectedCode: float64(4001),
				expectedMsg:  "Validation failed",
				expectedFields: []apperror.FieldError{
					{Field: "email", Message: "email is required"},
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				userService := new(mocks.MockUserService)
				mailerService := new(mocks.MockMailerService)
				handler := handlers.NewUserHandler(userService, mailerService)

				// Create a test context
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBufferString(tc.reqBody))

				// Call the handler
				handler.ForgotPassword(c)

				// Assert the response
				expectedBody := map[string]any{
					"code":    tc.expectedCode,
					"message": tc.expectedMsg,
					"fields":  tc.expectedFields,
				}
				var actualBody map[string]any
				_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, expectedBody["code"], actualBody["code"])
				assert.Equal(t, expectedBody["message"], actualBody["message"])
				assert.Equal(t, tc.expectedFields, utils.ToFieldErrors(actualBody["fields"]))

				// Assert mocks
				userService.AssertExpectations(t)
				mailerService.AssertExpectations(t)
			})
		}
	})

	t.Run("ForgotPassword - User Not Found", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"email": "notfound@example.com",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service method to return an error
		userService.On("ForgotPassword", mock.AnythingOfType("*dto.ForgotPasswordInput")).Return(&models.User{}, apperror.NewNotFoundError("User not found"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(body))

		// Call the handler
		handler.ForgotPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrNotFound),
			"message": "User not found",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ForgotPassword - Update User Error", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		requestBody := map[string]any{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(requestBody)

		// Mock the service methods
		userService.On("ForgotPassword", mock.AnythingOfType("*dto.ForgotPasswordInput")).Return(&models.User{}, apperror.NewDBUpdateError("Update failed"))

		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(body))

		// Call the handler
		handler.ForgotPassword(c)

		// Assert the response
		expectedBody := map[string]any{
			"code":    float64(apperror.ErrDBUpdate),
			"message": "Update failed",
		}
		var actualBody map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, expectedBody["code"], actualBody["code"])
		assert.Equal(t, expectedBody["message"], actualBody["message"])

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})

	t.Run("ForgotPassword - JSON Parse Error", func(t *testing.T) {
		userService := new(mocks.MockUserService)
		mailerService := new(mocks.MockMailerService)
		handler := handlers.NewUserHandler(userService, mailerService)

		// Create a test context with invalid JSON
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBufferString(`{invalid json}`))

		// Call the handler
		handler.ForgotPassword(c)

		// Assert the response - should return 400 for JSON parse error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Assert mocks
		userService.AssertExpectations(t)
		mailerService.AssertExpectations(t)
	})
}

package utils_test

import (
	stdErrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestRespondWith(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("RespondWithError_AppError", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		appErr := &apperror.AppError{
			HttpStatusCode: http.StatusBadRequest,
			Code:           1001,
			Message:        "App error occurred",
		}

		utils.RespondWithError(ctx, appErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		expectedJSON := `{"code":1001,"message":"App error occurred"}`
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})

	t.Run("RespondWithError_InternalServerError", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		internalErr := stdErrors.New("Internal server error occurred")

		utils.RespondWithError(ctx, internalErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		expectedJSON := `{"code":1000,"message":"Internal server error occurred"}`
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})

	t.Run("RespondWithError_GenericError", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		genericErr := apperror.NewInternalServerError("generic error message")

		utils.RespondWithError(ctx, genericErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		expectedJSON := `{"code":1000,"message":"generic error message"}`
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})

	t.Run("RespondWithOK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		body := gin.H{"success": true, "data": "some data"}

		utils.RespondWithOK(ctx, http.StatusOK, body)

		assert.Equal(t, http.StatusOK, w.Code)
		expectedJSON := `{"success":true,"data":"some data"}`
		assert.JSONEq(t, expectedJSON, w.Body.String())
	})
}

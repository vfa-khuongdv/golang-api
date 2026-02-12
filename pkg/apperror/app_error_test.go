package apperror_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAppError_Error(t *testing.T) {
	t.Run("without underlying error", func(t *testing.T) {
		// Arrange
		appErr := apperror.New(
			http.StatusInternalServerError,
			apperror.ErrInternalServer,
			"internal error",
		)
		expected := "code: 1000, message: internal error"

		// Act & Assert
		assert.Equal(t, expected, appErr.Error())
	})
}

func TestWrap(t *testing.T) {
	// Arrange
	underlying := apperror.New(
		http.StatusBadRequest,
		apperror.ErrBadRequest,
		"invalid input",
	)

	// Act
	appErr := apperror.Wrap(
		http.StatusBadRequest,
		apperror.ErrValidationFailed,
		"invalid request",
		underlying,
	)

	// Assert
	assert.NotNil(t, appErr)
	assert.Equal(t, apperror.ErrValidationFailed, appErr.Code)
	assert.Equal(t, "invalid request", appErr.Message)
}

func TestNew(t *testing.T) {
	// Act
	appErr := apperror.New(
		http.StatusUnauthorized,
		apperror.ErrUnauthorized,
		"unauthorized",
	)

	// Assert
	assert.NotNil(t, appErr)
	assert.Equal(t, apperror.ErrUnauthorized, appErr.Code)
	assert.Equal(t, "unauthorized", appErr.Message)
}

func TestIsAppError(t *testing.T) {
	t.Run("is AppError", func(t *testing.T) {
		// Arrange
		appErr := apperror.New(
			http.StatusForbidden,
			apperror.ErrForbidden,
			"forbidden",
		)

		// Assert
		assert.True(t, apperror.IsAppError(appErr))
	})

	t.Run("is not AppError", func(t *testing.T) {
		// Arrange
		err := assert.AnError

		// Assert
		assert.False(t, apperror.IsAppError(err))
	})
}

func TestToAppError(t *testing.T) {
	t.Run("is AppError", func(t *testing.T) {
		// Arrange
		appErr := apperror.New(
			http.StatusNotFound,
			apperror.ErrNotFound,
			"not found",
		)

		// Act
		result, ok := apperror.ToAppError(appErr)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, appErr, result)
	})

	t.Run("is not AppError", func(t *testing.T) {
		// Arrange
		err := assert.AnError

		// Act
		result, ok := apperror.ToAppError(err)

		// Assert
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestAppErrorWithUnderlyingError(t *testing.T) {
	// Arrange
	underlying := assert.AnError

	// Act
	appErr := apperror.Wrap(
		http.StatusInternalServerError,
		apperror.ErrInternalServer,
		"internal server error",
		underlying,
	)

	// Assert
	expected := "code: 1000, message: internal server error, error: " + underlying.Error()
	assert.Equal(t, expected, appErr.Error())
	assert.Equal(t, http.StatusInternalServerError, appErr.HttpStatusCode)
}

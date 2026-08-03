package apperror_test

import (
	"errors"
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

		// Act
		result := appErr.Error()

		// Assert
		assert.Equal(t, expected, result)
	})

	t.Run("with empty message", func(t *testing.T) {
		// Arrange
		appErr := apperror.New(http.StatusBadRequest, apperror.ErrBadRequest, "")

		// Act
		result := appErr.Error()

		// Assert
		assert.Equal(t, "code: 1002, message: ", result)
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
	assert.Equal(t, http.StatusBadRequest, appErr.HttpStatusCode)
	assert.Equal(t, underlying, appErr.Err)
}

func TestNew(t *testing.T) {
	// Arrange
	expectedCode := apperror.ErrUnauthorized
	expectedMessage := "unauthorized"

	// Act
	appErr := apperror.New(
		http.StatusUnauthorized,
		expectedCode,
		expectedMessage,
	)

	// Assert
	assert.NotNil(t, appErr)
	assert.Equal(t, expectedCode, appErr.Code)
	assert.Equal(t, expectedMessage, appErr.Message)
	assert.Equal(t, http.StatusUnauthorized, appErr.HttpStatusCode)
	assert.Nil(t, appErr.Err)
}

func TestIsAppError(t *testing.T) {
	t.Run("is AppError", func(t *testing.T) {
		// Arrange
		appErr := apperror.New(
			http.StatusForbidden,
			apperror.ErrForbidden,
			"forbidden",
		)

		// Act
		result := apperror.IsAppError(appErr)

		// Assert
		assert.True(t, result)
	})

	t.Run("is not AppError", func(t *testing.T) {
		// Arrange
		err := assert.AnError

		// Act
		result := apperror.IsAppError(err)

		// Assert
		assert.False(t, result)
	})

	t.Run("nil error", func(t *testing.T) {
		// Act
		result := apperror.IsAppError(nil)

		// Assert
		assert.False(t, result)
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

	t.Run("nil error", func(t *testing.T) {
		// Act
		result, ok := apperror.ToAppError(nil)

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

func TestUnwrap(t *testing.T) {
	// Arrange
	underlying := errors.New("wrapped error")
	appErr := apperror.Wrap(
		http.StatusBadRequest,
		apperror.ErrBadRequest,
		"bad request",
		underlying,
	)

	// Act & Assert
	assert.True(t, errors.Is(appErr, underlying), "errors.Is should find the wrapped error")
	assert.Equal(t, underlying, errors.Unwrap(appErr))
}

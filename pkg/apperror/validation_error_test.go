package apperror_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestValidateError(t *testing.T) {
	t.Run("NewValidationError formats correctly", func(t *testing.T) {
		// Arrange
		err := apperror.NewValidationError("Validation failed", []apperror.FieldError{
			{Field: "username", Message: "Username is required"},
			{Field: "email", Message: "Email is invalid"},
		})
		expected := "code: 4001, message: Validation failed, fields: [{username Username is required} {email Email is invalid}]"

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, expected, result)
	})

	t.Run("Wrap preserves fields and overrides message", func(t *testing.T) {
		// Arrange
		err := apperror.NewValidationError("Validation failed", []apperror.FieldError{
			{Field: "password", Message: "Password is too short"},
		})
		expected := "code: 4001, message: Wrapped validation error, fields: [{password Password is too short}]"

		// Act
		wrappedErr := err.Wrap(400, apperror.ErrValidationFailed, "Wrapped validation error")

		// Assert
		assert.NotNil(t, wrappedErr)
		assert.Equal(t, apperror.ErrValidationFailed, wrappedErr.Code)
		assert.Equal(t, "Wrapped validation error", wrappedErr.Message)
		assert.Equal(t, 1, len(wrappedErr.Fields))
		assert.Equal(t, expected, wrappedErr.Error())
	})

	t.Run("NewValidationError with multiple fields", func(t *testing.T) {
		// Arrange
		fieldErrors := []apperror.FieldError{
			{Field: "email", Message: "Email is required"},
			{Field: "password", Message: "Password must be at least 6 characters"},
		}

		// Act
		err := apperror.NewValidationError("Validation failed", fieldErrors)

		// Assert
		assert.Equal(t, apperror.ErrValidationFailed, err.Code)
		assert.Equal(t, "Validation failed", err.Message)
		assert.Equal(t, 2, len(err.Fields))

		for _, fieldError := range err.Fields {
			switch fieldError.Field {
			case "email":
				assert.Equal(t, "Email is required", fieldError.Message)
			case "password":
				assert.Equal(t, "Password must be at least 6 characters", fieldError.Message)
			}
		}
	})

	t.Run("NewValidationError with nil fields", func(t *testing.T) {
		// Act
		err := apperror.NewValidationError("Validation failed", nil)

		// Assert
		assert.Equal(t, apperror.ErrValidationFailed, err.Code)
		assert.Empty(t, err.Fields)
		assert.Equal(t, "code: 4001, message: Validation failed, fields: []", err.Error())
	})

	t.Run("NewValidationError with empty message", func(t *testing.T) {
		// Act
		err := apperror.NewValidationError("", []apperror.FieldError{
			{Field: "email", Message: "Email is required"},
		})

		// Assert
		assert.Equal(t, apperror.ErrValidationFailed, err.Code)
		assert.Equal(t, "", err.Message)
		assert.Equal(t, 1, len(err.Fields))
	})
}

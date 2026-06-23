package apperror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) *AppError
		wantCode int
		wantHTTP int
	}{
		// Generic errors
		{"InternalServerError", NewInternalServerError, ErrInternalServer, http.StatusInternalServerError},
		{"NotFoundError", NewNotFoundError, ErrNotFound, http.StatusNotFound},
		{"BadRequestError", NewBadRequestError, ErrBadRequest, http.StatusBadRequest},
		{"UnauthorizedError", NewUnauthorizedError, ErrUnauthorized, http.StatusUnauthorized},
		{"ForbiddenError", NewForbiddenError, ErrForbidden, http.StatusForbidden},
		{"ConflictError", NewConflictError, ErrConflict, http.StatusConflict},

		// Database errors
		{"DBConnectionError", NewDBConnectionError, ErrDBConnection, http.StatusInternalServerError},
		{"DBQueryError", NewDBQueryError, ErrDBQuery, http.StatusInternalServerError},
		{"DBInsertError", NewDBInsertError, ErrDBInsert, http.StatusInternalServerError},
		{"DBUpdateError", NewDBUpdateError, ErrDBUpdate, http.StatusInternalServerError},
		{"DBDeleteError", NewDBDeleteError, ErrDBDelete, http.StatusInternalServerError},

		// Cache errors
		{"CacheSetError", NewCacheSetError, ErrCacheSet, http.StatusInternalServerError},
		{"CacheGetError", NewCacheGetError, ErrCacheGet, http.StatusInternalServerError},
		{"CacheDeleteError", NewCacheDeleteError, ErrCacheDelete, http.StatusInternalServerError},
		{"CacheListError", NewCacheListError, ErrCacheList, http.StatusInternalServerError},
		{"CacheExistsError", NewCacheExistsError, ErrCacheExists, http.StatusInternalServerError},

		// Authentication errors
		{"TokenExpiredError", NewTokenExpiredError, ErrTokenExpired, http.StatusBadRequest},
		{"InvalidPasswordError", NewInvalidPasswordError, ErrInvalidPassword, http.StatusBadRequest},
		{"PasswordHashFailedError", NewPasswordHashFailedError, ErrPasswordHashFailed, http.StatusInternalServerError},
		{"PasswordMismatchError", NewPasswordMismatchError, ErrPasswordMismatch, http.StatusBadRequest},
		{"PasswordUnchangedError", NewPasswordUnchangedError, ErrPasswordUnchanged, http.StatusBadRequest},
		{"AccountLockedError", NewAccountLockedError, ErrAccountLocked, http.StatusTooManyRequests},

		// Common errors
		{"ParseError", NewParseError, ErrParseError, http.StatusBadRequest},
		{"ValidationDataError", NewValidationDataError, ErrValidationFailed, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msg := "test message"

			// Act
			err := tt.fn(msg)

			// Assert
			assert.Equal(t, tt.wantHTTP, err.HttpStatusCode, "HttpStatusCode")
			assert.Equal(t, tt.wantCode, err.Code, "Code")
			assert.Equal(t, msg, err.Message, "Message")
		})
	}
}

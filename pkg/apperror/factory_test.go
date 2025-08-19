package apperror

import (
	"net/http"
	"testing"
)

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) *AppError
		wantCode int
		wantHTTP int
	}{
		// Generic errors
		{"InternalError", NewInternalError, ErrInternal, http.StatusInternalServerError},
		{"NotFoundError", NewNotFoundError, ErrNotFound, http.StatusNotFound},
		{"BadRequestError", NewBadRequestError, ErrBadRequest, http.StatusBadRequest},
		{"UnauthorizedError", NewUnauthorizedError, ErrUnauthorized, http.StatusUnauthorized},
		{"ForbiddenError", NewForbiddenError, ErrForbidden, http.StatusForbidden},

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

		// Common errors
		{"ParseError", NewParseError, ErrParseError, http.StatusBadRequest},
		{"ValidationDataError", NewValidationDataError, ErrValidationFailed, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := "test message"
			err := tt.fn(msg)

			if err.HttpStatusCode != tt.wantHTTP {
				t.Errorf("expected HttpStatusCode %d, got %d", tt.wantHTTP, err.HttpStatusCode)
			}
			if err.Code != tt.wantCode {
				t.Errorf("expected Code %d, got %d", tt.wantCode, err.Code)
			}
			if err.Message != msg {
				t.Errorf("expected Message %s, got %s", msg, err.Message)
			}
		})
	}
}

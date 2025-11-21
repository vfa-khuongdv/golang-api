package apperror

import "net/http"

// === Generic errors ===
func NewInternalError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrInternal,
		Message:        message,
	}
}
func NewNotFoundError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusNotFound,
		Code:           ErrNotFound,
		Message:        message,
	}
}
func NewBadRequestError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrBadRequest,
		Message:        message,
	}
}
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusUnauthorized,
		Code:           ErrUnauthorized,
		Message:        message,
	}
}
func NewForbiddenError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusForbidden,
		Code:           ErrForbidden,
		Message:        message,
	}
}

// === Database errors ===
func NewDBConnectionError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrDBConnection,
		Message:        message,
	}
}

func NewDBQueryError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrDBQuery,
		Message:        message,
	}
}

func NewDBInsertError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrDBInsert,
		Message:        message,
	}
}

func NewDBUpdateError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrDBUpdate,
		Message:        message,
	}
}

func NewDBDeleteError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrDBDelete,
		Message:        message,
	}
}

// === Cache errors ===
func NewCacheSetError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrCacheSet,
		Message:        message,
	}
}
func NewCacheGetError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrCacheGet,
		Message:        message,
	}
}
func NewCacheDeleteError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrCacheDelete,
		Message:        message,
	}
}
func NewCacheListError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrCacheList,
		Message:        message,
	}
}
func NewCacheExistsError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrCacheExists,
		Message:        message,
	}
}

// === Authentication errors ===
func NewTokenExpiredError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrTokenExpired,
		Message:        message,
	}
}
func NewInvalidPasswordError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrInvalidPassword,
		Message:        message,
	}
}
func NewPasswordHashFailedError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrPasswordHashFailed,
		Message:        message,
	}
}
func NewPasswordMismatchError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrPasswordMismatch,
		Message:        message,
	}
}
func NewPasswordUnchangedError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrPasswordUnchanged,
		Message:        message,
	}
}

// === Common errors ===
func NewParseError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrParseError,
		Message:        message,
	}
}
func NewValidationDataError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrValidationFailed,
		Message:        message,
	}
}

// === MFA errors ===
func NewMfaAlreadyEnabledError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusConflict,
		Code:           ErrMfaAlreadyEnabled,
		Message:        message,
	}
}

func NewMfaNotEnabledError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrMfaNotEnabled,
		Message:        message,
	}
}

func NewMfaSetupNotInitiatedError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrMfaSetupNotInitiated,
		Message:        message,
	}
}

func NewMfaInvalidCodeError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrMfaInvalidCode,
		Message:        message,
	}
}

func NewMfaExpiredError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusBadRequest,
		Code:           ErrMfaExpired,
		Message:        message,
	}
}

func NewMfaSecretGenerationError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrMfaSecretGeneration,
		Message:        message,
	}
}

func NewMfaQRCodeGenerationError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrMfaQRCodeGeneration,
		Message:        message,
	}
}

func NewMfaBackupCodeError(message string) *AppError {
	return &AppError{
		HttpStatusCode: http.StatusInternalServerError,
		Code:           ErrMfaBackupCodeError,
		Message:        message,
	}
}

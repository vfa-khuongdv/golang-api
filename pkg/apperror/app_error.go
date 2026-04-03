package apperror

import "fmt"

// AppError represents a custom error with a code and message.
type AppError struct {
	HttpStatusCode int    `json:"-"`       // HTTP status code (optional)
	Code           int    `json:"code"`    // Error code
	Message        string `json:"message"` // Error message
	Err            error  `json:"-"`       // Underlying error (optional)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap creates a new AppError with an underlying error.
func Wrap(httpStatusCode, code int, message string, err error) *AppError {
	return &AppError{
		HttpStatusCode: httpStatusCode,
		Code:           code,
		Message:        message,
		Err:            err,
	}
}

// IsAppError checks if the provided error is of type AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// ToAppError converts an error to an AppError if it is of type AppError.
func ToAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// New creates a new AppError without an underlying error.
func New(httpStatusCode, code int, message string) *AppError {
	return &AppError{
		HttpStatusCode: httpStatusCode,
		Code:           code,
		Message:        message,
	}
}

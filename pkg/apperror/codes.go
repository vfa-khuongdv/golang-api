package apperror

const (
	// General errors
	ErrInternalServer = 1000 // Internal server error
	ErrNotFound       = 1001 // Resource not found
	ErrBadRequest     = 1002 // Invalid or bad request
	ErrUnauthorized   = 1003 // Unauthorized access
	ErrForbidden      = 1004 // Forbidden access
	ErrConflict       = 1005 // Conflict error

	// Database errors
	ErrDBConnection = 2000 // Failed to connect to DB
	ErrDBQuery      = 2001 // DB query error
	ErrDBInsert     = 2002 // DB insert error
	ErrDBUpdate     = 2003 // DB update error
	ErrDBDelete     = 2004 // DB delete error

	// Authentication errors
	ErrTokenExpired       = 3001 // Token has expired
	ErrInvalidPassword    = 3002 // Invalid password
	ErrPasswordHashFailed = 3003 // Failed to hash password
	ErrPasswordMismatch   = 3004 // Password mismatch
	ErrPasswordUnchanged  = 3005 // Old and new password are the same

	// Common
	ErrParseError       = 4000 // Parsing or field error
	ErrValidationFailed = 4001 // Validation failed
	ErrEmptyData        = 4007 // No data provided

	// Cache errors
	ErrCacheSet    = 4002 // Set cache error
	ErrCacheGet    = 4003 // Get cache error
	ErrCacheDelete = 4004 // Delete cache error
	ErrCacheList   = 4005 // List cache error
	ErrCacheExists = 4006 // Cache key exists check error
)

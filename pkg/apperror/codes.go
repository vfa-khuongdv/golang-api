package apperror

const (
	// General errors
	ErrInternal     = 1000 // Internal server error
	ErrNotFound     = 1001 // Resource not found
	ErrBadRequest   = 1002 // Invalid or bad request
	ErrUnauthorized = 3000 // Unauthorized access
	ErrForbidden    = 3001 // Forbidden access

	// Database errors
	ErrDBConnection = 2000 // Failed to connect to DB
	ErrDBQuery      = 2001 // DB query error
	ErrDBInsert     = 2002 // DB insert error
	ErrDBUpdate     = 2003 // DB update error
	ErrDBDelete     = 2004 // DB delete error

	// Authentication errors
	ErrTokenExpired       = 3002 // Token has expired
	ErrInvalidPassword    = 3003 // Invalid password
	ErrPasswordHashFailed = 3004 // Failed to hash password
	ErrPasswordMismatch   = 3005 // Password mismatch
	ErrPasswordUnchanged  = 3006 // Old and new password are the same

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

	// MFA errors
	ErrMfaAlreadyEnabled    = 5000 // MFA is already enabled
	ErrMfaNotEnabled        = 5001 // MFA is not enabled
	ErrMfaSetupNotInitiated = 5002 // MFA setup not initiated
	ErrMfaInvalidCode       = 5003 // Invalid MFA code
	ErrMfaExpired           = 5004 // MFA session expired
	ErrMfaSecretGeneration  = 5005 // Failed to generate MFA secret
	ErrMfaQRCodeGeneration  = 5006 // Failed to generate QR code
	ErrMfaBackupCodeError   = 5007 // Backup code error
)

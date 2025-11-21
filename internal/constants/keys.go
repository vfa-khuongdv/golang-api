package constants

// LIMIT is the maximum number of items to be returned in a single page
const LIMIT int = 50

// MFA constants
const (
	MFA_CODE_EXPIRY       int    = 5 * 60 // 5 minutes in seconds
	MFA_BACKUP_CODE_COUNT int    = 10
	MFA_ISSUER            string = "GolangCMS"
)

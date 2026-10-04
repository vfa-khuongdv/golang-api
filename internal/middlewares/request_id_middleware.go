package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

const (
	// RequestIDHeader is the HTTP header name for the request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for storing the request ID
	RequestIDKey = "RequestID"
)

// RequestIDMiddleware adds a unique request ID to each request
// If the client provides an X-Request-ID header, it will be used
// Otherwise, a new UUID will be generated
// The request ID is:
// - Stored in the Gin context for use by handlers
// - Added to the response header
// - Injected into ctx.Request.Context() for automatic logging
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// A client ID is only kept when it is safe to log and echo back:
		// otherwise a header with a newline or quotes could forge log lines.
		requestID := c.GetHeader(RequestIDHeader)
		if !isSafeRequestID(requestID) {
			requestID = uuid.New().String()
		}

		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		req := c.Request.WithContext(logger.WithRequestIDContext(c.Request.Context(), requestID))
		c.Request = req

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the Gin context
// Returns empty string if request ID is not found
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// maxRequestIDLength bounds a client-provided request ID.
const maxRequestIDLength = 128

// isSafeRequestID reports whether id is non-empty, at most maxRequestIDLength
// long, and made only of letters, digits and -_.:= (enough for UUIDs and load
// balancer trace IDs such as "Root=1-...").
func isSafeRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune("-_.:=", r):
		default:
			return false
		}
	}
	return true
}

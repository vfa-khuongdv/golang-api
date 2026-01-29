package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// - Available for logging and tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client provided a request ID
		requestID := c.GetHeader(RequestIDHeader)

		// If no request ID provided, generate a new UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context for handlers to access
		c.Set(RequestIDKey, requestID)

		// Add request ID to response header
		c.Writer.Header().Set(RequestIDHeader, requestID)

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

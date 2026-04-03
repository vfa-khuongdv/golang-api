package middlewares

import (
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
		requestID := c.GetHeader(RequestIDHeader)

		if requestID == "" {
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

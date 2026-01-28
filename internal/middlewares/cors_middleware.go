package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
// Security: Configure CORS_ALLOWED_ORIGINS environment variable with specific
// origins (e.g., "http://localhost:3000,https://example.com"). Never use "*"
// in production with credentials enabled.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := utils.GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

		// Check if origin is allowed and set appropriate headers
		if isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the request origin is in the allowed origins list
func isOriginAllowed(origin, allowedOrigins string) bool {
	if origin == "" {
		return false
	}

	if allowedOrigins == "*" {
		return true
	}

	origins := strings.SplitSeq(allowedOrigins, ",")
	for allowed := range origins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}

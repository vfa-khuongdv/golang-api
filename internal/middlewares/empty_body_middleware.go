package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// Middleware to reject requests with empty JSON body
func EmptyBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			var bodyBytes []byte
			var err error

			// Check if body is not nil before reading
			if c.Request.Body != nil {
				bodyBytes, err = io.ReadAll(c.Request.Body)
			}

			// Check if body is nil, error reading, or empty content
			if c.Request.Body == nil || err != nil || len(bytes.TrimSpace(bodyBytes)) == 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    apperror.ErrEmptyData,
					"message": "Request body cannot be empty",
				})
				return
			}
			// Replace the body so the handler can read it again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		c.Next()
	}
}

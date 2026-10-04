package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// BodySizeLimit caps the request body at maxBytes, so a client cannot make the
// server read an unbounded body into memory. A declared Content-Length over
// the limit is answered with 413 at once; a body without one (chunked) fails
// when read past the limit, which binding reports as an invalid body. It must
// run before any middleware that reads the body.
func BodySizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.ContentLength > maxBytes {
			utils.RespondWithError(ctx, apperror.New(http.StatusRequestEntityTooLarge, apperror.ErrPayloadTooLarge, "Request body is too large"))
			return
		}
		// Body is nil for a request built in code without one (e.g. in tests).
		if ctx.Request.Body != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)
		}
		ctx.Next()
	}
}

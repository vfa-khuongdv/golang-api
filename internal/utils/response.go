package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// RespondWithError sends a JSON error response with the given status code and error
// Parameters:
//   - ctx: Gin context for the request
//   - statusCode: HTTP status code to return
//   - err: Error to be serialized as JSON response body
//   - 1. If the error is a ValidationError, it includes validation error code, message, and fields.
//   - 2. If the error is an AppError, it includes application error code and message.
//   - 3. If the error is neither, it returns a generic internal error response.
func RespondWithError(ctx *gin.Context, err error) {
	// 1. If the error is a ValidationError, return its code, message, and fields
	if validateErr, ok := err.(*apperror.ValidationError); ok {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"code":    validateErr.Code,
				"message": validateErr.Message,
				"fields":  validateErr.Fields,
			},
		)
		return
	}

	// 2. If the error is an AppError, return its code and message
	if appErr, ok := err.(*apperror.AppError); ok {
		ctx.AbortWithStatusJSON(
			appErr.HttpStatusCode,
			gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		)
		return
	}
	// 3. If the error is not a ValidationError or AppError, return a generic internal error
	ctx.AbortWithStatusJSON(
		http.StatusInternalServerError,
		gin.H{
			"code":    apperror.ErrInternalServer,
			"message": err.Error(),
		},
	)
}

// RespondWithOK sends a JSON response with the given status code and body
// Parameters:
//   - ctx: Gin context for the request
//   - statusCode: HTTP status code to return
//   - body: Data to be serialized as JSON response body
func RespondWithOK(ctx *gin.Context, statusCode int, body any) {
	ctx.AbortWithStatusJSON(statusCode, body)
}

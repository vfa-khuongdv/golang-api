package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// AuthMiddleware is a Gin middleware function that handles JWT authentication
// It validates the Authorization header and extracts the JWT token
// The middleware checks if:
// - Authorization header exists and has "Bearer " prefix
// - Token is valid and can be parsed
// - Token has "access" scope
// If validation succeeds, it sets the user ID from token claims in context
// If validation fails, it returns 401 Unauthorized
func AuthMiddleware() gin.HandlerFunc {
	jwtService := services.NewJWTService()
	return func(ctx *gin.Context) {

		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondWithError(ctx, apperror.NewUnauthorizedError("Authorization header required"))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtService.ValidateTokenWithScope(tokenString, services.TokenScopeAccess)
		if err != nil {
			utils.RespondWithError(ctx, apperror.NewUnauthorizedError("Unauthorized"))
			return
		}

		ctx.Set("UserID", claims.ID)
		ctx.Next()
	}
}

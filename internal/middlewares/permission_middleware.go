package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

// RequirePermission allows the request only when one of the authenticated
// user's roles grants the permission. It must run after AuthMiddleware, which
// sets "UserID". Permissions are read from the database on every request, so a
// change takes effect immediately.
// Responds 401 without a user, 403 without the permission, 500 on lookup errors.
func RequirePermission(roleService services.RoleService, permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := ctx.Get("UserID")
		if !ok {
			utils.RespondWithError(ctx, apperror.NewUnauthorizedError("Unauthorized"))
			return
		}

		allowed, err := roleService.HasPermission(ctx.Request.Context(), userID.(uint), permission)
		if err != nil {
			logger.WithContext(ctx.Request.Context()).Errorf("Permission check failed: %v", err)
			utils.RespondWithError(ctx, err)
			return
		}
		if !allowed {
			utils.RespondWithError(ctx, apperror.NewForbiddenError("You do not have permission to perform this action"))
			return
		}

		ctx.Next()
	}
}

package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUserIDFromContext(ctx *gin.Context) (uint, error) {
	userIdInterface, exists := ctx.Get("UserID")
	if !exists {
		return 0, errors.New("User ID not found in context")
	}

	userId, ok := userIdInterface.(uint)
	if !ok {
		return 0, errors.New("User ID in context has invalid type")
	}

	return userId, nil
}

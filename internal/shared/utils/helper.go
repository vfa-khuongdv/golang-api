package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrUserIDNotFound    = errors.New("user ID not found in context")
	ErrUserIDInvalidType = errors.New("user ID in context has invalid type")
)

func GetUserIDFromContext(ctx *gin.Context) (uint, error) {
	userIdInterface, exists := ctx.Get("UserID")
	if !exists {
		return 0, ErrUserIDNotFound
	}

	userId, ok := userIdInterface.(uint)
	if !ok {
		return 0, ErrUserIDInvalidType
	}

	return userId, nil
}

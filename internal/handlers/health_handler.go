package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

func HealthCheck(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "healthy"})
}

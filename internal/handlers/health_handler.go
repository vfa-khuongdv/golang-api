package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

var startTime = time.Now()

const (
	APIVersion = "1.0.0"
)

func HealthCheck(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "healthy"})
}

func VersionInfo(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{
		"version":    APIVersion,
		"build_time": startTime.Format(time.RFC3339),
		"uptime":     time.Since(startTime).String(),
	})
}

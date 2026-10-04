package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

var startTime = time.Now()

const (
	APIVersion = "1.0.0"

	// readinessTimeout keeps the probe answering well within a load balancer's
	// check timeout even when the database hangs.
	readinessTimeout = 2 * time.Second
)

func HealthCheck(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "healthy"})
}

// ReadinessCheck reports whether the service can serve traffic: it answers 503
// when ping cannot reach the database, so the load balancer stops routing here.
func ReadinessCheck(ping func(context.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), readinessTimeout)
		defer cancel()

		if err := ping(pingCtx); err != nil {
			logger.WithContext(ctx.Request.Context()).Warnf("Readiness check failed: %v", err)
			utils.RespondWithError(ctx, apperror.Wrap(http.StatusServiceUnavailable, apperror.ErrDBConnection, "Database is unavailable", err))
			return
		}
		utils.RespondWithOK(ctx, http.StatusOK, gin.H{"status": "ready"})
	}
}

func VersionInfo(ctx *gin.Context) {
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{
		"version":    APIVersion,
		"build_time": startTime.Format(time.RFC3339),
		"uptime":     time.Since(startTime).String(),
	})
}

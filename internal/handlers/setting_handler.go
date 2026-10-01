package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type settingHandlerImpl struct {
	settingService services.SettingService
}

func NewSettingHandler(settingService services.SettingService) *settingHandlerImpl {
	return &settingHandlerImpl{settingService: settingService}
}

func (handler *settingHandlerImpl) GetSettings(ctx *gin.Context) {
	settings, err := handler.settingService.GetSettings(ctx.Request.Context())
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Get settings failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, settings)
}

func (handler *settingHandlerImpl) UpdateSettings(ctx *gin.Context) {
	var input dto.UpdateSettingsInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		validateError := utils.TranslateValidationErrors(err, input)
		utils.RespondWithError(ctx, validateError)
		return
	}

	if err := handler.settingService.UpdateSettings(ctx.Request.Context(), &input); err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Update settings failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}

	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Update settings successfully"})
}

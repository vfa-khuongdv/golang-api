package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type roleHandlerImpl struct {
	roleService services.RoleService
}

func NewRoleHandler(roleService services.RoleService) *roleHandlerImpl {
	return &roleHandlerImpl{roleService: roleService}
}

// parseIDParam reads the :id path parameter as a positive integer. On failure
// it writes a 400 response and returns false.
func parseIDParam(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		utils.RespondWithError(ctx, apperror.NewBadRequestError("Invalid id"))
		return 0, false
	}
	return uint(id), true
}

func (handler *roleHandlerImpl) ListRoles(ctx *gin.Context) {
	roles, err := handler.roleService.ListRoles(ctx.Request.Context())
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("List roles failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, nonNil(roles))
}

func (handler *roleHandlerImpl) GetRole(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	role, err := handler.roleService.GetRole(ctx.Request.Context(), id)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Get role failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, role)
}

func (handler *roleHandlerImpl) CreateRole(ctx *gin.Context) {
	var input dto.RoleInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(ctx, utils.TranslateValidationErrors(err, input))
		return
	}
	role, err := handler.roleService.CreateRole(ctx.Request.Context(), &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Create role failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusCreated, role)
}

func (handler *roleHandlerImpl) UpdateRole(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var input dto.RoleInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(ctx, utils.TranslateValidationErrors(err, input))
		return
	}
	role, err := handler.roleService.UpdateRole(ctx.Request.Context(), id, &input)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Update role failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, role)
}

func (handler *roleHandlerImpl) DeleteRole(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	if err := handler.roleService.DeleteRole(ctx.Request.Context(), id); err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Delete role failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, gin.H{"message": "Delete role successfully"})
}

func (handler *roleHandlerImpl) ListPermissions(ctx *gin.Context) {
	perms, err := handler.roleService.ListPermissions(ctx.Request.Context())
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("List permissions failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, nonNil(perms))
}

func (handler *roleHandlerImpl) SetUserRoles(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var input dto.SetUserRolesInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(ctx, utils.TranslateValidationErrors(err, input))
		return
	}
	roles, err := handler.roleService.SetUserRoles(ctx.Request.Context(), id, input.RoleIDs)
	if err != nil {
		logger.WithContext(ctx.Request.Context()).Errorf("Set user roles failed: %v", err)
		utils.RespondWithError(ctx, err)
		return
	}
	utils.RespondWithOK(ctx, http.StatusOK, nonNil(roles))
}

// nonNil makes a nil slice serialize as [] instead of null.
func nonNil[T models.Role | models.Permission](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

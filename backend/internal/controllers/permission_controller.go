package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type PermissionController struct {
	permissionService *services.PermissionService
}

func NewPermissionController(permissionService *services.PermissionService) *PermissionController {
	return &PermissionController{
		permissionService: permissionService,
	}
}

// GrantPermission godoc
// @Summary      Выдача права
// @Description  Выдаёт право должности или сотруднику.
// @Description  Требует право "permissions.grant" в соответствующем scope.
// @Description  Можно выдать только те права, которые есть у самого выдающего.
// @Description  Основатели могут выдавать любые права.
// @Tags         Права
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.GrantPermissionRequest true "Данные для выдачи права"
// @Success      200 {object} dto.MessageResponse "Право успешно выдано"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      403 {object} dto.ErrorResponse "Нет права на выдачу прав или попытка выдать право, которого нет у самого пользователя"
// @Failure      404 {object} dto.ErrorResponse "Разрешение не найдено"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/permissions/grant [post]
func (ctrl *PermissionController) GrantPermission(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req dto.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация: должен быть указан либо position_id, либо employee_id
	if req.PositionID == nil && req.EmployeeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either position_id or employee_id must be provided"})
		return
	}

	if req.PositionID != nil && req.EmployeeID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot specify both position_id and employee_id"})
		return
	}

	// Выдаём право с проверкой
	err = ctrl.permissionService.GrantPermissionWithCheck(
		userID,
		orgID,
		req.PositionID,
		req.EmployeeID,
		req.PermissionCode,
		req.ScopeType,
		req.ScopeID,
	)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "permission granted successfully"})
}

// RevokePermission godoc
// @Summary      Отзыв права
// @Description  Отзывает право у должности или сотрудника.
// @Description  Требует право "permissions.revoke" в соответствующем scope.
// @Description  Основатели могут отзывать любые права.
// @Tags         Права
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.RevokePermissionRequest true "Данные для отзыва права"
// @Success      200 {object} dto.MessageResponse "Право успешно отозвано"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      403 {object} dto.ErrorResponse "Нет права на отзыв прав"
// @Failure      404 {object} dto.ErrorResponse "Разрешение не найдено"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/permissions/revoke [post]
func (ctrl *PermissionController) RevokePermission(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req dto.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация
	if req.PositionID == nil && req.EmployeeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either position_id or employee_id must be provided"})
		return
	}

	if req.PositionID != nil && req.EmployeeID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot specify both position_id and employee_id"})
		return
	}

	// Отзываём право с проверкой
	err = ctrl.permissionService.RevokePermissionWithCheck(
		userID,
		orgID,
		req.PositionID,
		req.EmployeeID,
		req.PermissionCode,
		req.ScopeType,
		req.ScopeID,
	)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "permission revoked successfully"})
}

// GetPositionPermissions godoc
// @Summary      Получить права должности
// @Description  Возвращает список всех прав, выданных должности.
// @Description  Требует право "permissions.view".
// @Tags         Права
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        posId path int true "ID должности" minimum(1) example(1)
// @Success      200 {object} dto.PermissionGrantsListResponse "Список прав должности"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      403 {object} dto.ErrorResponse "Нет права на просмотр прав"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/positions/{posId}/permissions [get]
func (ctrl *PermissionController) GetPositionPermissions(c *gin.Context) {
	posID, err := strconv.ParseInt(c.Param("posId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	grants, err := ctrl.permissionService.GetPositionPermissions(posID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.PermissionGrantsListResponse{
		Grants: dto.ToPositionPermissionGrantResponses(grants),
	})
}

// GetEmployeePermissions godoc
// @Summary      Получить права сотрудника
// @Description  Возвращает список всех индивидуальных прав сотрудника.
// @Description  Требует право "permissions.view".
// @Tags         Права
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        empId path int true "ID сотрудника" minimum(1) example(1)
// @Success      200 {object} dto.PermissionGrantsListResponse "Список прав сотрудника"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      403 {object} dto.ErrorResponse "Нет права на просмотр прав"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/employees/{empId}/permissions [get]
func (ctrl *PermissionController) GetEmployeePermissions(c *gin.Context) {
	empID, err := strconv.ParseInt(c.Param("empId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
		return
	}

	grants, err := ctrl.permissionService.GetEmployeePermissionGrants(empID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.PermissionGrantsListResponse{
		Grants: dto.ToEmployeePermissionGrantResponses(grants),
	})
}

// GetAllPermissions godoc
// @Summary      Получить все доступные права
// @Description  Возвращает список всех прав в системе.
// @Description  Требует доступ к организации.
// @Tags         Права
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {object} object{permissions=[]models.Permission} "Список всех прав"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/permissions [get]
func (ctrl *PermissionController) GetAllPermissions(c *gin.Context) {
	permissions, err := ctrl.permissionService.GetAllPermissions()
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"permissions": permissions})
}

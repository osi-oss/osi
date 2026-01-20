package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type OrganizationController struct {
	orgService *services.OrganizationService
}

func NewOrganizationController(orgService *services.OrganizationService) *OrganizationController {
	return &OrganizationController{orgService: orgService}
}

// CreateOrganization создает новую организацию
// @Summary      Создание организации
// @Description  Создаёт новую организацию. Пользователь становится основателем.
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateOrganizationRequest true "Данные организации"
// @Success      201  {object}  dto.OrganizationResponse  "Организация создана"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Router       /organizations [post]
func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := ctrl.orgService.CreateOrganization(userID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToOrganizationResponse(org))
}

// GetOrganization получает организацию по ID
// @Summary      Получение организации
// @Description  Возвращает организацию по ID
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Success      200  {object}  dto.OrganizationResponse  "Организация"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет доступа"
// @Failure      404  {object}  map[string]string  "Организация не найдена"
// @Router       /organizations/{orgId} [get]
func (ctrl *OrganizationController) GetOrganization(c *gin.Context) {
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

	org, err := ctrl.orgService.GetOrganization(orgID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToOrganizationResponse(org))
}

// GetUserOrganizations получает все организации пользователя
// @Summary      Список организаций пользователя
// @Description  Возвращает все организации, в которых состоит пользователь
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string][]dto.OrganizationResponse  "Список организаций"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Router       /organizations [get]
func (ctrl *OrganizationController) GetUserOrganizations(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgs, err := ctrl.orgService.GetUserOrganizations(userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"organizations": dto.ToOrganizationResponses(orgs)})
}

// UpdateOrganization обновляет организацию
// @Summary      Обновление организации
// @Description  Обновляет данные организации
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        request body dto.UpdateOrganizationRequest true "Новые данные"
// @Success      200  {object}  dto.OrganizationResponse  "Организация обновлена"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет доступа"
// @Failure      404  {object}  map[string]string  "Организация не найдена"
// @Router       /organizations/{orgId} [put]
func (ctrl *OrganizationController) UpdateOrganization(c *gin.Context) {
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

	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := ctrl.orgService.UpdateOrganization(orgID, userID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToOrganizationResponse(org))
}

// DeleteOrganization удаляет организацию
// @Summary      Удаление организации
// @Description  Удаляет организацию. Только для основателей.
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Success      200  {object}  map[string]string  "Организация удалена"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет доступа"
// @Failure      404  {object}  map[string]string  "Организация не найдена"
// @Router       /organizations/{orgId} [delete]
func (ctrl *OrganizationController) DeleteOrganization(c *gin.Context) {
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

	if err := ctrl.orgService.DeleteOrganization(orgID, userID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "organization deleted successfully"})
}

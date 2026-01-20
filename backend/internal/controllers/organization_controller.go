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

// CreateOrganization godoc
// @Summary      Создание новой организации
// @Description  Создаёт новую организацию в системе.
// @Description  Текущий пользователь автоматически становится основателем (founder) организации.
// @Description  Основатель имеет полные права на управление организацией.
// @Description  Новая организация создаётся со статусом "draft".
// @Tags         Организации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateOrganizationRequest true "Данные для создания организации"
// @Success      201 {object} dto.OrganizationResponse "Организация успешно создана"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: название обязательно"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// GetOrganization godoc
// @Summary      Получение организации по ID
// @Description  Возвращает полную информацию об организации.
// @Description  Доступ имеют только участники организации (члены или основатели).
// @Tags         Организации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {object} dto.OrganizationResponse "Данные организации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет доступа к организации"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// GetUserOrganizations godoc
// @Summary      Список организаций текущего пользователя
// @Description  Возвращает все организации, в которых пользователь является участником или основателем.
// @Description  Включает организации со всеми статусами (draft, pending, approved, rejected).
// @Tags         Организации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.OrganizationsListResponse "Список организаций пользователя"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// UpdateOrganization godoc
// @Summary      Обновление данных организации
// @Description  Обновляет информацию об организации.
// @Description  Можно обновлять только те поля, которые переданы в запросе.
// @Description  Требует доступа к организации (основатель или участник с соответствующими правами).
// @Tags         Организации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.UpdateOrganizationRequest true "Новые данные организации"
// @Success      200 {object} dto.OrganizationResponse "Организация успешно обновлена"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации данных"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на редактирование организации"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// DeleteOrganization godoc
// @Summary      Удаление организации
// @Description  Полностью удаляет организацию из системы.
// @Description  Доступно только основателям организации.
// @Description  **ВНИМАНИЕ**: Удаление безвозвратно удаляет все связанные данные:
// @Description  - Все локации организации
// @Description  - Все отделы и позиции
// @Description  - Все данные об участниках и сотрудниках
// @Tags         Организации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID удаляемой организации" minimum(1) example(1)
// @Success      200 {object} dto.MessageResponse "Организация успешно удалена"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Только основатель может удалить организацию"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

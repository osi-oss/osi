package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type EmployeeController struct {
	orgService *services.OrganizationService
}

func NewEmployeeController(orgService *services.OrganizationService) *EmployeeController {
	return &EmployeeController{
		orgService: orgService,
	}
}

// GetMyOrganizations godoc
// @Summary      Получить мои организации
// @Description  Возвращает список организаций, где текущий пользователь работает или является основателем.
// @Description  Для каждой организации возвращается статус пользователя и дополнительная информация.
// @Tags         Employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.MyOrganizationInfo "Список организаций"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /employees/my-organizations [get]
func (ctrl *EmployeeController) GetMyOrganizations(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgs, err := ctrl.orgService.GetUserOrganizationsWithDetails(userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, orgs)
}

// GetOrganizationEmployees godoc
// @Summary      Получить всех сотрудников организации
// @Description  Возвращает список всех сотрудников организации с подробной информацией.
// @Description  Доступно только для членов организации.
// @Tags         Employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {array} dto.EmployeeDetailResponse "Список сотрудников"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет доступа к организации"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/employees [get]
func (ctrl *EmployeeController) GetOrganizationEmployees(c *gin.Context) {
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

	// Проверяем доступ к организации
	_, err = ctrl.orgService.GetOrganization(orgID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	employees, err := ctrl.orgService.GetOrganizationEmployees(orgID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, employees)
}

// GetOrganizationHierarchy godoc
// @Summary      Получить иерархию организации
// @Description  Возвращает полную иерархию организации: локации, отделы, должности, сотрудники.
// @Description  Доступно только для членов организации.
// @Tags         Employees
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {object} dto.OrganizationHierarchyResponse "Иерархия организации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет доступа к организации"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/hierarchy [get]
func (ctrl *EmployeeController) GetOrganizationHierarchy(c *gin.Context) {
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

	// Проверяем доступ к организации
	_, err = ctrl.orgService.GetOrganization(orgID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	hierarchy, err := ctrl.orgService.GetOrganizationHierarchy(orgID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, hierarchy)
}

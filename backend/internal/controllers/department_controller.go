package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type DepartmentController struct {
	departmentService *services.DepartmentService
}

func NewDepartmentController(departmentService *services.DepartmentService) *DepartmentController {
	return &DepartmentController{departmentService: departmentService}
}

// CreateDepartment godoc
// @Summary      Создание нового отдела
// @Description  Создаёт новый отдел в указанной локации.
// @Description  Отделы могут иметь иерархическую структуру (указывается parent_id).
// @Description  Требуется право "departments.create" или статус основателя.
// @Tags         Отделы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        request body dto.CreateDepartmentRequest true "Данные для создания отдела"
// @Success      201 {object} dto.DepartmentResponse "Отдел успешно создан"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: name обязательно"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права departments.create"
// @Failure      404 {object} dto.ErrorResponse "Локация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId}/departments [post]
func (ctrl *DepartmentController) CreateDepartment(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	department, err := ctrl.departmentService.CreateDepartment(locationID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToDepartmentResponse(department))
}

// GetLocationDepartments godoc
// @Summary      Список отделов локации
// @Description  Возвращает все отделы указанной локации.
// @Description  Включает как корневые, так и вложенные отделы.
// @Description  Требуется доступ к организации.
// @Tags         Отделы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Success      200 {object} dto.DepartmentsListResponse "Список отделов локации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId}/departments [get]
func (ctrl *DepartmentController) GetLocationDepartments(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	departments, err := ctrl.departmentService.GetLocationDepartments(locationID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"departments": dto.ToDepartmentResponses(departments)})
}

// GetDepartment godoc
// @Summary      Получение отдела по ID
// @Description  Возвращает полную информацию об отделе.
// @Description  Требуется доступ к организации.
// @Tags         Отделы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        deptId path int true "ID отдела" minimum(1) example(1)
// @Success      200 {object} dto.DepartmentResponse "Данные отдела"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      404 {object} dto.ErrorResponse "Отдел не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId}/departments/{deptId} [get]
func (ctrl *DepartmentController) GetDepartment(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("deptId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	department, err := ctrl.departmentService.GetDepartment(deptID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToDepartmentResponse(department))
}

// UpdateDepartment godoc
// @Summary      Обновление отдела
// @Description  Обновляет данные отдела.
// @Description  Можно изменить название, описание и родительский отдел.
// @Description  Требуется право "departments.update" или статус основателя.
// @Tags         Отделы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        deptId path int true "ID отдела" minimum(1) example(1)
// @Param        request body dto.UpdateDepartmentRequest true "Новые данные отдела"
// @Success      200 {object} dto.DepartmentResponse "Отдел успешно обновлён"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации данных"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права departments.update"
// @Failure      404 {object} dto.ErrorResponse "Отдел не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId}/departments/{deptId} [put]
func (ctrl *DepartmentController) UpdateDepartment(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("deptId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	department, err := ctrl.departmentService.UpdateDepartment(deptID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToDepartmentResponse(department))
}

// DeleteDepartment godoc
// @Summary      Удаление отдела
// @Description  Удаляет отдел из локации.
// @Description  **ВНИМАНИЕ**: Удаление отдела удалит все дочерние отделы.
// @Description  Требуется право "departments.delete" или статус основателя.
// @Tags         Отделы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        deptId path int true "ID отдела" minimum(1) example(1)
// @Success      200 {object} dto.MessageResponse "Отдел успешно удалён"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права departments.delete"
// @Failure      404 {object} dto.ErrorResponse "Отдел не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId}/departments/{deptId} [delete]
func (ctrl *DepartmentController) DeleteDepartment(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("deptId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	if err := ctrl.departmentService.DeleteDepartment(deptID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "department deleted successfully"})
}

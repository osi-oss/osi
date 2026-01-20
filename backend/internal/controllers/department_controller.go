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

// CreateDepartment создает новый отдел
// @Summary      Создание отдела
// @Description  Создаёт новый отдел в локации
// @Tags         departments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        request body dto.CreateDepartmentRequest true "Данные отдела"
// @Success      201  {object}  dto.DepartmentResponse  "Отдел создан"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права departments.create"
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

// GetLocationDepartments получает все отделы локации
// @Summary      Список отделов
// @Description  Возвращает все отделы локации
// @Tags         departments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Success      200  {object}  map[string][]dto.DepartmentResponse  "Список отделов"
// @Failure      401  {object}  map[string]string  "Не авторизован"
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

// GetDepartment получает отдел по ID
// @Summary      Получение отдела
// @Description  Возвращает отдел по ID
// @Tags         departments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        deptId path int true "ID отдела"
// @Success      200  {object}  dto.DepartmentResponse  "Отдел"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      404  {object}  map[string]string  "Отдел не найден"
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

// UpdateDepartment обновляет отдел
// @Summary      Обновление отдела
// @Description  Обновляет данные отдела
// @Tags         departments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        deptId path int true "ID отдела"
// @Param        request body dto.UpdateDepartmentRequest true "Новые данные"
// @Success      200  {object}  dto.DepartmentResponse  "Отдел обновлён"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права departments.update"
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

// DeleteDepartment удаляет отдел
// @Summary      Удаление отдела
// @Description  Удаляет отдел из локации
// @Tags         departments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        deptId path int true "ID отдела"
// @Success      200  {object}  map[string]string  "Отдел удалён"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права departments.delete"
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

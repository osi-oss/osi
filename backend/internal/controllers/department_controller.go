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
func (ctrl *DepartmentController) CreateDepartment(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	locationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	department, err := ctrl.departmentService.CreateDepartment(locationID, userID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToDepartmentResponse(department))
}

// GetLocationDepartments получает все отделы локации
func (ctrl *DepartmentController) GetLocationDepartments(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	locationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	departments, err := ctrl.departmentService.GetLocationDepartments(locationID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"departments": dto.ToDepartmentResponses(departments)})
}

// GetDepartment получает отдел по ID
func (ctrl *DepartmentController) GetDepartment(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	deptID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	department, err := ctrl.departmentService.GetDepartment(deptID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToDepartmentResponse(department))
}

// UpdateDepartment обновляет отдел
func (ctrl *DepartmentController) UpdateDepartment(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	deptID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	department, err := ctrl.departmentService.UpdateDepartment(deptID, userID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToDepartmentResponse(department))
}

// DeleteDepartment удаляет отдел
func (ctrl *DepartmentController) DeleteDepartment(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	deptID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	if err := ctrl.departmentService.DeleteDepartment(deptID, userID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "department deleted successfully"})
}

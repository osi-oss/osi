package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/services"
)

type DepartmentController struct {
	departmentService *services.DepartmentService
}

func NewDepartmentController(departmentService *services.DepartmentService) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
	}
}

// CreateDepartment создает новый отдел в локации
func (ctrl *DepartmentController) CreateDepartment(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var input services.CreateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	department, err := ctrl.departmentService.CreateDepartment(locationID, userID.(int64), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrLocationNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          department.ID,
		"location_id": department.LocationID,
		"parent_id":   department.ParentID,
		"name":        department.Name,
		"description": department.Description,
		"created_at":  department.CreatedAt,
	})
}

// GetLocationDepartments получает все отделы локации
func (ctrl *DepartmentController) GetLocationDepartments(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	departments, err := ctrl.departmentService.GetLocationDepartments(locationID, userID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrLocationNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(departments))
	for i, dept := range departments {
		result[i] = gin.H{
			"id":          dept.ID,
			"location_id": dept.LocationID,
			"parent_id":   dept.ParentID,
			"name":        dept.Name,
			"description": dept.Description,
			"created_at":  dept.CreatedAt,
			"updated_at":  dept.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"departments": result})
}

// GetDepartment получает конкретный отдел по ID
func (ctrl *DepartmentController) GetDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	department, err := ctrl.departmentService.GetDepartment(departmentID, userID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          department.ID,
		"location_id": department.LocationID,
		"parent_id":   department.ParentID,
		"name":        department.Name,
		"description": department.Description,
		"created_at":  department.CreatedAt,
		"updated_at":  department.UpdatedAt,
	})
}

// UpdateDepartment обновляет отдел
func (ctrl *DepartmentController) UpdateDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	var input services.UpdateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	department, err := ctrl.departmentService.UpdateDepartment(departmentID, userID.(int64), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          department.ID,
		"location_id": department.LocationID,
		"parent_id":   department.ParentID,
		"name":        department.Name,
		"description": department.Description,
		"created_at":  department.CreatedAt,
		"updated_at":  department.UpdatedAt,
	})
}

// DeleteDepartment удаляет отдел
func (ctrl *DepartmentController) DeleteDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := ctrl.departmentService.DeleteDepartment(departmentID, userID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department deleted successfully"})
}

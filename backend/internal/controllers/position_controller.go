package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/services"
)

type PositionController struct {
	positionService *services.PositionService
}

func NewPositionController(positionService *services.PositionService) *PositionController {
	return &PositionController{
		positionService: positionService,
	}
}

// CreatePosition создает новую позицию в организации
func (ctrl *PositionController) CreatePosition(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var input services.CreatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	position, err := ctrl.positionService.CreatePosition(orgID, userID.(int64), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrOrganizationNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              position.ID,
		"organization_id": position.OrganizationID,
		"department_id":   position.DepartmentID,
		"name":            position.Name,
		"is_admin":        position.IsAdmin,
		"description":     position.Description,
		"created_at":      position.CreatedAt,
	})
}

// GetOrganizationPositions получает все позиции организации
func (ctrl *PositionController) GetOrganizationPositions(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	positions, err := ctrl.positionService.GetOrganizationPositions(orgID, userID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrOrganizationNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(positions))
	for i, pos := range positions {
		result[i] = gin.H{
			"id":              pos.ID,
			"organization_id": pos.OrganizationID,
			"department_id":   pos.DepartmentID,
			"name":            pos.Name,
			"is_admin":        pos.IsAdmin,
			"description":     pos.Description,
			"created_at":      pos.CreatedAt,
			"updated_at":      pos.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"positions": result})
}

// GetDepartmentPositions получает все позиции отдела
func (ctrl *PositionController) GetDepartmentPositions(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	positions, err := ctrl.positionService.GetDepartmentPositions(deptID, userID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(positions))
	for i, pos := range positions {
		result[i] = gin.H{
			"id":              pos.ID,
			"organization_id": pos.OrganizationID,
			"department_id":   pos.DepartmentID,
			"name":            pos.Name,
			"is_admin":        pos.IsAdmin,
			"description":     pos.Description,
			"created_at":      pos.CreatedAt,
			"updated_at":      pos.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"positions": result})
}

// GetPosition получает конкретную позицию по ID
func (ctrl *PositionController) GetPosition(c *gin.Context) {
	positionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	position, err := ctrl.positionService.GetPosition(positionID, userID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrPositionNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              position.ID,
		"organization_id": position.OrganizationID,
		"department_id":   position.DepartmentID,
		"name":            position.Name,
		"is_admin":        position.IsAdmin,
		"description":     position.Description,
		"created_at":      position.CreatedAt,
		"updated_at":      position.UpdatedAt,
	})
}

// UpdatePosition обновляет позицию
func (ctrl *PositionController) UpdatePosition(c *gin.Context) {
	positionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	var input services.UpdatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	position, err := ctrl.positionService.UpdatePosition(positionID, userID.(int64), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrPositionNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              position.ID,
		"organization_id": position.OrganizationID,
		"department_id":   position.DepartmentID,
		"name":            position.Name,
		"is_admin":        position.IsAdmin,
		"description":     position.Description,
		"created_at":      position.CreatedAt,
		"updated_at":      position.UpdatedAt,
	})
}

// DeletePosition удаляет позицию
func (ctrl *PositionController) DeletePosition(c *gin.Context) {
	positionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := ctrl.positionService.DeletePosition(positionID, userID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrPositionNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, services.ErrUnauthorized):
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "position deleted successfully"})
}

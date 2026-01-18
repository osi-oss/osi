package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type PositionController struct {
	positionService *services.PositionService
}

func NewPositionController(positionService *services.PositionService) *PositionController {
	return &PositionController{positionService: positionService}
}

// CreatePosition создает новую позицию
func (ctrl *PositionController) CreatePosition(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req dto.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	position, err := ctrl.positionService.CreatePosition(orgID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToPositionResponse(position))
}

// GetOrganizationPositions получает все позиции организации
func (ctrl *PositionController) GetOrganizationPositions(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	positions, err := ctrl.positionService.GetOrganizationPositions(orgID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"positions": dto.ToPositionResponses(positions)})
}

// GetDepartmentPositions получает все позиции отдела
func (ctrl *PositionController) GetDepartmentPositions(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("deptId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}

	positions, err := ctrl.positionService.GetDepartmentPositions(deptID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"positions": dto.ToPositionResponses(positions)})
}

// GetPosition получает позицию по ID
func (ctrl *PositionController) GetPosition(c *gin.Context) {
	posID, err := strconv.ParseInt(c.Param("posId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	position, err := ctrl.positionService.GetPosition(posID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToPositionResponse(position))
}

// UpdatePosition обновляет позицию
func (ctrl *PositionController) UpdatePosition(c *gin.Context) {
	posID, err := strconv.ParseInt(c.Param("posId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	var req dto.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	position, err := ctrl.positionService.UpdatePosition(posID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToPositionResponse(position))
}

// DeletePosition удаляет позицию
func (ctrl *PositionController) DeletePosition(c *gin.Context) {
	posID, err := strconv.ParseInt(c.Param("posId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}

	if err := ctrl.positionService.DeletePosition(posID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "position deleted successfully"})
}

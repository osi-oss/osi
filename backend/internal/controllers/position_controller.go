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
// @Summary      Создание позиции
// @Description  Создаёт новую позицию в организации
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        request body dto.CreatePositionRequest true "Данные позиции"
// @Success      201  {object}  dto.PositionResponse  "Позиция создана"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права positions.create"
// @Router       /organizations/{orgId}/positions [post]
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
// @Summary      Список позиций организации
// @Description  Возвращает все позиции организации
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Success      200  {object}  map[string][]dto.PositionResponse  "Список позиций"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Router       /organizations/{orgId}/positions [get]
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
// @Summary      Список позиций отдела
// @Description  Возвращает все позиции конкретного отдела
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        deptId path int true "ID отдела"
// @Success      200  {object}  map[string][]dto.PositionResponse  "Список позиций"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Router       /organizations/{orgId}/locations/{locId}/departments/{deptId}/positions [get]
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
// @Summary      Получение позиции
// @Description  Возвращает позицию по ID
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        posId path int true "ID позиции"
// @Success      200  {object}  dto.PositionResponse  "Позиция"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      404  {object}  map[string]string  "Позиция не найдена"
// @Router       /organizations/{orgId}/positions/{posId} [get]
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
// @Summary      Обновление позиции
// @Description  Обновляет данные позиции
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        posId path int true "ID позиции"
// @Param        request body dto.UpdatePositionRequest true "Новые данные"
// @Success      200  {object}  dto.PositionResponse  "Позиция обновлена"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права positions.update"
// @Router       /organizations/{orgId}/positions/{posId} [put]
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
// @Summary      Удаление позиции
// @Description  Удаляет позицию из организации
// @Tags         positions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        posId path int true "ID позиции"
// @Success      200  {object}  map[string]string  "Позиция удалена"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права positions.delete"
// @Router       /organizations/{orgId}/positions/{posId} [delete]
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

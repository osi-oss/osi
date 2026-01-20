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

// CreatePosition godoc
// @Summary      Создание новой должности
// @Description  Создаёт новую должность в организации.
// @Description  Можно привязать к отделу (department_id) или оставить общей для организации.
// @Description  Административные позиции (is_admin=true) имеют расширенные права.
// @Description  Требуется право "positions.create" или статус основателя.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.CreatePositionRequest true "Данные для создания должности"
// @Success      201 {object} dto.PositionResponse "Должность успешно создана"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: name обязательно"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права positions.create"
// @Failure      404 {object} dto.ErrorResponse "Организация или отдел не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// GetOrganizationPositions godoc
// @Summary      Список всех должностей организации
// @Description  Возвращает все должности организации.
// @Description  Включает как привязанные к отделам, так и общие позиции.
// @Description  Требуется доступ к организации.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {object} dto.PositionsListResponse "Список должностей организации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// GetDepartmentPositions godoc
// @Summary      Список должностей отдела
// @Description  Возвращает все должности, привязанные к конкретному отделу.
// @Description  Требуется доступ к организации.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        deptId path int true "ID отдела" minimum(1) example(1)
// @Success      200 {object} dto.PositionsListResponse "Список должностей отдела"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// GetPosition godoc
// @Summary      Получение должности по ID
// @Description  Возвращает полную информацию о должности.
// @Description  Требуется доступ к организации.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        posId path int true "ID должности" minimum(1) example(1)
// @Success      200 {object} dto.PositionResponse "Данные должности"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      404 {object} dto.ErrorResponse "Должность не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// UpdatePosition godoc
// @Summary      Обновление должности
// @Description  Обновляет данные должности.
// @Description  Можно изменить название, описание, привязку к отделу и административный статус.
// @Description  Требуется право "positions.update" или статус основателя.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        posId path int true "ID должности" minimum(1) example(1)
// @Param        request body dto.UpdatePositionRequest true "Новые данные должности"
// @Success      200 {object} dto.PositionResponse "Должность успешно обновлена"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации данных"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права positions.update"
// @Failure      404 {object} dto.ErrorResponse "Должность не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

// DeletePosition godoc
// @Summary      Удаление должности
// @Description  Удаляет должность из организации.
// @Description  **ВНИМАНИЕ**: Нельзя удалить должность, если на ней есть сотрудники.
// @Description  Требуется право "positions.delete" или статус основателя.
// @Tags         Должности
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        posId path int true "ID должности" minimum(1) example(1)
// @Success      200 {object} dto.MessageResponse "Должность успешно удалена"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права positions.delete"
// @Failure      404 {object} dto.ErrorResponse "Должность не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
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

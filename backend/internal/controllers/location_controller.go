package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type LocationController struct {
	locationService *services.LocationService
}

func NewLocationController(locationService *services.LocationService) *LocationController {
	return &LocationController{locationService: locationService}
}

// CreateLocation godoc
// @Summary      Создание новой локации
// @Description  Создаёт новую локацию (офис, филиал, склад и т.д.) в организации.
// @Description  Требуется право "locations.create" или статус основателя.
// @Description  Поле source указывает источник данных: manual (вручную), egrul (из ЕГРЮЛ), api (через API).
// @Tags         Локации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.CreateLocationRequest true "Данные для создания локации"
// @Success      201 {object} dto.LocationResponse "Локация успешно создана"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: name и source обязательны"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права locations.create"
// @Failure      404 {object} dto.ErrorResponse "Организация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations [post]
func (ctrl *LocationController) CreateLocation(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req dto.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, err := ctrl.locationService.CreateLocation(orgID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToLocationResponse(location))
}

// GetOrganizationLocations godoc
// @Summary      Список локаций организации
// @Description  Возвращает все локации организации (активные и неактивные).
// @Description  Требуется доступ к организации (участник или основатель).
// @Tags         Локации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {object} dto.LocationsListResponse "Список локаций организации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет доступа к организации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations [get]
func (ctrl *LocationController) GetOrganizationLocations(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	locations, err := ctrl.locationService.GetOrganizationLocations(orgID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"locations": dto.ToLocationResponses(locations)})
}

// GetLocation godoc
// @Summary      Получение локации по ID
// @Description  Возвращает полную информацию о локации.
// @Description  Требуется доступ к организации.
// @Tags         Локации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Success      200 {object} dto.LocationResponse "Данные локации"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      404 {object} dto.ErrorResponse "Локация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId} [get]
func (ctrl *LocationController) GetLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	location, err := ctrl.locationService.GetLocation(locationID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToLocationResponse(location))
}

// UpdateLocation godoc
// @Summary      Обновление локации
// @Description  Обновляет данные локации.
// @Description  Можно обновить название, адрес, источник, статус верификации и активность.
// @Description  Требуется право "locations.update" или статус основателя.
// @Tags         Локации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Param        request body dto.UpdateLocationRequest true "Новые данные локации"
// @Success      200 {object} dto.LocationResponse "Локация успешно обновлена"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации данных"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права locations.update"
// @Failure      404 {object} dto.ErrorResponse "Локация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId} [put]
func (ctrl *LocationController) UpdateLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, err := ctrl.locationService.UpdateLocation(locationID, req)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToLocationResponse(location))
}

// DeleteLocation godoc
// @Summary      Удаление локации
// @Description  Удаляет локацию из организации.
// @Description  **ВНИМАНИЕ**: Удаление локации удалит все связанные отделы.
// @Description  Требуется право "locations.delete" или статус основателя.
// @Tags         Локации
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        locId path int true "ID локации" minimum(1) example(1)
// @Success      200 {object} dto.MessageResponse "Локация успешно удалена"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет права locations.delete"
// @Failure      404 {object} dto.ErrorResponse "Локация не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/locations/{locId} [delete]
func (ctrl *LocationController) DeleteLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	if err := ctrl.locationService.DeleteLocation(locationID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "location deleted successfully"})
}

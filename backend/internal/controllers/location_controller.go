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

// CreateLocation создает новую локацию
// @Summary      Создание локации
// @Description  Создаёт новую локацию в организации
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        request body dto.CreateLocationRequest true "Данные локации"
// @Success      201  {object}  dto.LocationResponse  "Локация создана"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права locations.create"
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

// GetOrganizationLocations получает все локации организации
// @Summary      Список локаций
// @Description  Возвращает все локации организации
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Success      200  {object}  map[string][]dto.LocationResponse  "Список локаций"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет доступа"
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

// GetLocation получает локацию по ID
// @Summary      Получение локации
// @Description  Возвращает локацию по ID
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Success      200  {object}  dto.LocationResponse  "Локация"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      404  {object}  map[string]string  "Локация не найдена"
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

// UpdateLocation обновляет локацию
// @Summary      Обновление локации
// @Description  Обновляет данные локации
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Param        request body dto.UpdateLocationRequest true "Новые данные"
// @Success      200  {object}  dto.LocationResponse  "Локация обновлена"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права locations.update"
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

// DeleteLocation удаляет локацию
// @Summary      Удаление локации
// @Description  Удаляет локацию из организации
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации"
// @Param        locId path int true "ID локации"
// @Success      200  {object}  map[string]string  "Локация удалена"
// @Failure      401  {object}  map[string]string  "Не авторизован"
// @Failure      403  {object}  map[string]string  "Нет права locations.delete"
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

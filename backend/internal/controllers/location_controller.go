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

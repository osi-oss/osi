package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/services"
)

type LocationController struct {
	locationService *services.LocationService
}

func NewLocationController(locationService *services.LocationService) *LocationController {
	return &LocationController{
		locationService: locationService,
	}
}

// CreateLocation создает новый филиал/локацию для организации
func (ctrl *LocationController) CreateLocation(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var input services.CreateLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	location, err := ctrl.locationService.CreateLocation(orgID, userID.(int64), input)
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
		"id":              location.ID,
		"organization_id": location.OrganizationID,
		"name":            location.Name,
		"address":         location.Address,
		"source":          location.Source,
		"is_verified":     location.IsVerified,
		"is_active":       location.IsActive,
		"created_at":      location.CreatedAt,
	})
}

// GetOrganizationLocations получает все локации организации
func (ctrl *LocationController) GetOrganizationLocations(c *gin.Context) {
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

	locations, err := ctrl.locationService.GetOrganizationLocations(orgID, userID.(int64))
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

	result := make([]gin.H, len(locations))
	for i, loc := range locations {
		result[i] = gin.H{
			"id":              loc.ID,
			"organization_id": loc.OrganizationID,
			"name":            loc.Name,
			"address":         loc.Address,
			"source":          loc.Source,
			"is_verified":     loc.IsVerified,
			"is_active":       loc.IsActive,
			"created_at":      loc.CreatedAt,
			"updated_at":      loc.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"locations": result})
}

// GetLocation получает конкретную локацию по ID
func (ctrl *LocationController) GetLocation(c *gin.Context) {
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

	location, err := ctrl.locationService.GetLocation(locationID, userID.(int64))
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

	c.JSON(http.StatusOK, gin.H{
		"id":              location.ID,
		"organization_id": location.OrganizationID,
		"name":            location.Name,
		"address":         location.Address,
		"source":          location.Source,
		"is_verified":     location.IsVerified,
		"is_active":       location.IsActive,
		"created_at":      location.CreatedAt,
		"updated_at":      location.UpdatedAt,
	})
}

// UpdateLocation обновляет локацию
func (ctrl *LocationController) UpdateLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location id"})
		return
	}

	var input services.UpdateLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	location, err := ctrl.locationService.UpdateLocation(locationID, userID.(int64), input)
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

	c.JSON(http.StatusOK, gin.H{
		"id":              location.ID,
		"organization_id": location.OrganizationID,
		"name":            location.Name,
		"address":         location.Address,
		"source":          location.Source,
		"is_verified":     location.IsVerified,
		"is_active":       location.IsActive,
		"created_at":      location.CreatedAt,
		"updated_at":      location.UpdatedAt,
	})
}

// DeleteLocation удаляет локацию
func (ctrl *LocationController) DeleteLocation(c *gin.Context) {
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

	if err := ctrl.locationService.DeleteLocation(locationID, userID.(int64)); err != nil {
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

	c.JSON(http.StatusOK, gin.H{"message": "location deleted successfully"})
}

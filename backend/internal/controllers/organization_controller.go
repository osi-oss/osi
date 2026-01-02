package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/services"
)

type OrganizationController struct {
	orgService *services.OrganizationService
}

func NewOrganizationController(orgService *services.OrganizationService) *OrganizationController {
	return &OrganizationController{
		orgService: orgService,
	}
}

// CreateOrganization создает новую организацию
func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) {
	var input services.CreateOrganizationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	org, err := ctrl.orgService.CreateOrganization(userID.(int64), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            org.ID,
		"name":          org.Name,
		"legal_name":    org.LegalName,
		"inn":           org.INN,
		"ogrn":          org.OGRN,
		"kpp":           org.KPP,
		"legal_address": org.LegalAddress,
		"status":        org.Status,
		"created_at":    org.CreatedAt,
	})
}

// GetOrganization получает организацию по ID
func (ctrl *OrganizationController) GetOrganization(c *gin.Context) {
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

	org, err := ctrl.orgService.GetOrganization(orgID, userID.(int64))
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

	c.JSON(http.StatusOK, gin.H{
		"id":            org.ID,
		"name":          org.Name,
		"legal_name":    org.LegalName,
		"inn":           org.INN,
		"ogrn":          org.OGRN,
		"kpp":           org.KPP,
		"legal_address": org.LegalAddress,
		"status":        org.Status,
		"created_at":    org.CreatedAt,
		"updated_at":    org.UpdatedAt,
	})
}

// GetUserOrganizations получает все организации пользователя
func (ctrl *OrganizationController) GetUserOrganizations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	orgs, err := ctrl.orgService.GetUserOrganizations(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, len(orgs))
	for i, org := range orgs {
		response[i] = gin.H{
			"id":            org.ID,
			"name":          org.Name,
			"legal_name":    org.LegalName,
			"inn":           org.INN,
			"ogrn":          org.OGRN,
			"kpp":           org.KPP,
			"legal_address": org.LegalAddress,
			"status":        org.Status,
			"created_at":    org.CreatedAt,
			"updated_at":    org.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"organizations": response})
}

// UpdateOrganization обновляет организацию
func (ctrl *OrganizationController) UpdateOrganization(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var input services.CreateOrganizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	org, err := ctrl.orgService.UpdateOrganization(orgID, userID.(int64), input)
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

	c.JSON(http.StatusOK, gin.H{
		"id":            org.ID,
		"name":          org.Name,
		"legal_name":    org.LegalName,
		"inn":           org.INN,
		"ogrn":          org.OGRN,
		"kpp":           org.KPP,
		"legal_address": org.LegalAddress,
		"status":        org.Status,
		"created_at":    org.CreatedAt,
		"updated_at":    org.UpdatedAt,
	})
}

// DeleteOrganization удаляет организацию
func (ctrl *OrganizationController) DeleteOrganization(c *gin.Context) {
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

	err = ctrl.orgService.DeleteOrganization(orgID, userID.(int64))
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

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

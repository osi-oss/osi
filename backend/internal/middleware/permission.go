package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/services"
)

// PermissionMiddleware handles permission checks
type PermissionMiddleware struct {
	permissionSvc *services.PermissionService
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(permissionSvc *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionSvc: permissionSvc,
	}
}

// parseOrgID extracts organization ID from URL parameter
func parseOrgID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("orgId"), 10, 64)
}

// parseScopeID extracts scope ID from URL parameter based on scope type
func parseScopeID(c *gin.Context, scopeType models.ScopeType) *int64 {
	var paramName string
	switch scopeType {
	case models.ScopeLocation:
		paramName = "locId"
	case models.ScopeDepartment:
		paramName = "deptId"
	default:
		return nil
	}

	idStr := c.Param(paramName)
	if idStr == "" {
		return nil
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

// RequireOrgAccess checks if user has access to organization (is founder or active member)
func (m *PermissionMiddleware) RequireOrgAccess(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	orgID, err := parseOrgID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		c.Abort()
		return
	}

	// Store orgID in context for later use
	c.Set("orgID", orgID)

	hasAccess, err := m.permissionSvc.UserHasAccessToOrganization(userID.(int64), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access"})
		c.Abort()
		return
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		c.Abort()
		return
	}

	c.Next()
}

// RequireScopedPermissionHierarchy checks permission with full scope hierarchy support
// This middleware validates that user has the required permission considering
// the full context hierarchy (organization → location → department)
func (m *PermissionMiddleware) RequireScopedPermissionHierarchy(
	permissionCode string,
	requestedScope models.ScopeType,
	extractContext func(*gin.Context) *models.PermissionContext,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		orgID, err := parseOrgID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
			c.Abort()
			return
		}

		c.Set("orgID", orgID)

		// Check if user has access to organization first
		hasAccess, err := m.permissionSvc.UserHasAccessToOrganization(userID.(int64), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access"})
			c.Abort()
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		// Extract full context hierarchy
		context := extractContext(c)

		// Check permission with hierarchy
		hasPermission, err := m.permissionSvc.UserHasScopedPermissionWithHierarchy(
			userID.(int64),
			orgID,
			permissionCode,
			requestedScope,
			context,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}

		// Store context in request for controller use
		c.Set("permissionContext", context)
		c.Next()
	}
}

// Helper functions to extract context from URL parameters

// extractOrganizationContext extracts organization-level context
func extractOrganizationContext(c *gin.Context) *models.PermissionContext {
	return &models.PermissionContext{}
}

// extractLocationContext extracts location-level context
func extractLocationContext(c *gin.Context) *models.PermissionContext {
	locID, _ := strconv.ParseInt(c.Param("locId"), 10, 64)
	return &models.PermissionContext{
		LocationID: &locID,
	}
}

// extractDepartmentContext extracts department-level context (includes location)
func extractDepartmentContext(c *gin.Context) *models.PermissionContext {
	locID, _ := strconv.ParseInt(c.Param("locId"), 10, 64)
	deptID, _ := strconv.ParseInt(c.Param("deptId"), 10, 64)
	return &models.PermissionContext{
		LocationID:   &locID,
		DepartmentID: &deptID,
	}
}

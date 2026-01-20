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

// RequirePermission returns a middleware that checks org-level permission (no scope)
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return m.RequireScopedPermission(permissionCode, models.ScopeOrganization)
}

// RequireScopedPermission returns a middleware that checks scoped permission
// scopeType determines which URL parameter to use for scope_id:
// - ScopeOrganization: no scope_id needed
// - ScopeLocation: uses :locId from URL
// - ScopeDepartment: uses :deptId from URL
func (m *PermissionMiddleware) RequireScopedPermission(permissionCode string, scopeType models.ScopeType) gin.HandlerFunc {
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

		// Store orgID in context for later use
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

		// Parse scope ID from URL
		scopeID := parseScopeID(c, scopeType)

		// Check scoped permission
		hasPermission, err := m.permissionSvc.UserHasScopedPermission(userID.(int64), orgID, permissionCode, scopeType, scopeID)
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

		c.Next()
	}
}

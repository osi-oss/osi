package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// RequirePermission returns a middleware that checks if user has specific permission
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
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

		// Check specific permission
		hasPermission, err := m.permissionSvc.UserHasPermission(userID.(int64), orgID, permissionCode)
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

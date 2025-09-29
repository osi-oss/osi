package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidateJSON проверяет что Content-Type application/json
func ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

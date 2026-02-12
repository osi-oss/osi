package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/validators"
)

// SecurityMiddleware проверяет запросы на потенциальные атаки
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем URL параметры
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if validators.ContainsSQLInjection(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid characters detected in request",
						"field": key,
					})
					c.Abort()
					return
				}

				if validators.ContainsXSS(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid characters detected in request",
						"field": key,
					})
					c.Abort()
					return
				}
			}
		}

		// Проверяем заголовки на подозрительное содержимое
		suspiciousHeaders := []string{"Referer", "X-Forwarded-For"}
		for _, header := range suspiciousHeaders {
			value := c.GetHeader(header)
			if validators.ContainsSQLInjection(value) || validators.ContainsXSS(value) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "Invalid characters detected in headers",
					"header": header,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

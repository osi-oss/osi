package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/osi-oss/osi/internal/models"
)

// AuthRequired проверяет JWT токен из куки или Authorization header
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Пытаемся получить токен из куки
		if cookie, err := c.Cookie("auth_token"); err == nil {
			tokenString = cookie
		} else {
			// Если куки нет, пытаемся получить из Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
				c.Abort()
				return
			}
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Парсим и валидируем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Извлекаем данные из токена
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// JWT хранит числа как float64, конвертируем в int64
			if sub, ok := claims["sub"].(float64); ok {
				c.Set("userID", int64(sub))
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			// Email
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			// Status пользователя
			if status, ok := claims["status"].(string); ok {
				c.Set("status", status)
			} else {
				// Для совместимости со старыми токенами
				c.Set("status", string(models.UserStatusActive))
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/models"
)

// RequireStatus проверяет, что статус пользователя соответствует одному из требуемых
func RequireStatus(required ...models.UserStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.GetString("status")
		if status == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "User status not found in token",
				"message": "Please re-authenticate",
			})
			c.Abort()
			return
		}

		currentStatus := models.UserStatus(status)
		for _, r := range required {
			if currentStatus == r {
				c.Next()
				return
			}
		}

		// Формируем понятное сообщение об ошибке
		var message string
		var nextStep string
		switch currentStatus {
		case models.UserStatusPendingEmail:
			message = "Email verification required"
			nextStep = "verify_email"
		case models.UserStatusPendingProfile:
			message = "Profile completion required"
			nextStep = "complete_profile"
		default:
			message = "Access denied"
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":          "Insufficient user status",
			"message":        message,
			"current_status": status,
			"next_step":      nextStep,
		})
		c.Abort()
	}
}

// RequireActiveUser требует полностью активированного пользователя
func RequireActiveUser() gin.HandlerFunc {
	return RequireStatus(models.UserStatusActive)
}

// RequireEmailVerified требует подтверждённый email (pending_profile или active)
func RequireEmailVerified() gin.HandlerFunc {
	return RequireStatus(models.UserStatusPendingProfile, models.UserStatusActive)
}

// RequirePendingProfile требует статус pending_profile (для заполнения профиля)
func RequirePendingProfile() gin.HandlerFunc {
	return RequireStatus(models.UserStatusPendingProfile)
}

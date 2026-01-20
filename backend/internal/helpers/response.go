package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/apperrors"
)

// RespondError отправляет ошибку клиенту с правильным HTTP кодом
func RespondError(c *gin.Context, err error) {
	code := apperrors.GetHTTPCode(err)
	message := apperrors.GetMessage(err)
	c.JSON(code, gin.H{"error": message})
}

// RespondSuccess отправляет успешный ответ
func RespondSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

// RespondCreated отправляет ответ о создании
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(201, data)
}

// RespondOK отправляет успешный ответ 200
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(200, data)
}

// RespondMessage отправляет сообщение
func RespondMessage(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"message": message})
}

package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/apperrors"
)

// GetUserID извлекает ID пользователя из контекста
func GetUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, apperrors.ErrUnauthorized
	}

	id, ok := userID.(int64)
	if !ok {
		return 0, apperrors.ErrUnauthorized
	}

	return id, nil
}

// MustGetUserID извлекает ID пользователя или паникует
func MustGetUserID(c *gin.Context) int64 {
	id, err := GetUserID(c)
	if err != nil {
		panic("userID not found in context - AuthRequired middleware missing?")
	}
	return id
}

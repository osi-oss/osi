package repository

import (
	"time"

	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create создает новый токен для сброса пароля
func (r *PasswordResetRepository) Create(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

// GetByToken находит токен по строковому значению
func (r *PasswordResetRepository) GetByToken(tokenStr string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.Where("token = ? AND used = ? AND expires_at > ?",
		tokenStr, false, time.Now()).First(&token).Error
	return &token, err
}

// MarkAsUsed помечает токен как использованный
func (r *PasswordResetRepository) MarkAsUsed(tokenID uint) error {
	return r.db.Model(&models.PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("used", true).Error
}

// DeleteExpired удаляет просроченные токены
func (r *PasswordResetRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordResetToken{}).Error
}

// DeleteByUserID удаляет все токены пользователя (при смене пароля)
func (r *PasswordResetRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).
		Delete(&models.PasswordResetToken{}).Error
}

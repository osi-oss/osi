package repository

import (
	"time"

	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type AuthCodeRepository struct {
	db *gorm.DB
}

func NewAuthCodeRepository(db *gorm.DB) *AuthCodeRepository {
	return &AuthCodeRepository{db: db}
}

// Create создаёт новый код аутентификации
func (r *AuthCodeRepository) Create(code *models.AuthCode) error {
	return r.db.Create(code).Error
}

// GetLatestByUserID получает последний активный код пользователя
func (r *AuthCodeRepository) GetLatestByUserID(userID int64) (*models.AuthCode, error) {
	var code models.AuthCode
	err := r.db.Where("user_id = ? AND used = ? AND expires_at > ?",
		userID, false, time.Now()).
		Order("created_at DESC").
		First(&code).Error
	return &code, err
}

// GetValidCode проверяет код и возвращает его если он валиден
func (r *AuthCodeRepository) GetValidCode(userID int64, codeStr string) (*models.AuthCode, error) {
	var code models.AuthCode
	err := r.db.Where("user_id = ? AND code = ? AND used = ? AND expires_at > ? AND attempts < ?",
		userID, codeStr, false, time.Now(), models.MaxAuthCodeAttempts).
		First(&code).Error
	return &code, err
}

// IncrementAttempts увеличивает счётчик попыток
func (r *AuthCodeRepository) IncrementAttempts(codeID int64) error {
	return r.db.Model(&models.AuthCode{}).
		Where("id = ?", codeID).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

// MarkAsUsed помечает код как использованный
func (r *AuthCodeRepository) MarkAsUsed(codeID int64) error {
	return r.db.Model(&models.AuthCode{}).
		Where("id = ?", codeID).
		Update("used", true).Error
}

// DeleteByUserID удаляет все коды пользователя
func (r *AuthCodeRepository) DeleteByUserID(userID int64) error {
	return r.db.Where("user_id = ?", userID).
		Delete(&models.AuthCode{}).Error
}

// DeleteExpired удаляет все истёкшие коды
func (r *AuthCodeRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.AuthCode{}).Error
}

// CountRecentByUserID подсчитывает количество кодов за последние N минут
func (r *AuthCodeRepository) CountRecentByUserID(userID int64, minutes int) (int64, error) {
	var count int64
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.db.Model(&models.AuthCode{}).
		Where("user_id = ? AND created_at > ?", userID, since).
		Count(&count).Error
	return count, err
}

// InvalidateAllForUser помечает все активные коды пользователя как использованные
func (r *AuthCodeRepository) InvalidateAllForUser(userID int64) error {
	return r.db.Model(&models.AuthCode{}).
		Where("user_id = ? AND used = ?", userID, false).
		Update("used", true).Error
}

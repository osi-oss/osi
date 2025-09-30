package db

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

func SyncDb(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Organisation{},
		&models.PasswordResetToken{}, // Добавляем новую модель
	)
}

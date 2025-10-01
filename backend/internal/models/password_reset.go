package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken токен для восстановления пароля
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;size:255"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Связь с пользователем
	User User `gorm:"foreignKey:UserID"`
}

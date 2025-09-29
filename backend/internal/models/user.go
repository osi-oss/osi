package models

import "time"

// User представляет пользователя системы
type User struct {
	ID            uint   `gorm:"primaryKey"`
	FirstName     string `gorm:"size:100"`
	LastName      string `gorm:"size:100"`
	MiddleName    string `gorm:"size:100"`
	Email         string `gorm:"uniqueIndex"`
	Phone         string `gorm:"uniqueIndex"`
	Password      string `gorm:"size:255"`
	Verified      bool   `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Organisations []Organisation `gorm:"many2many:user_organisations;"`
}

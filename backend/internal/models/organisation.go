package models

import "time"

type Organisation struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255"`
	LegalName string `gorm:"size:255"`
	INN       string `gorm:"size:20"`
	Status    string `gorm:"size:50"` // "pending", "approved", "rejected"
	CreatedAt time.Time
	UpdatedAt time.Time
}

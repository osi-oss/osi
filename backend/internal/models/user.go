package models

import "time"

type User struct {
	ID            uint   `gorm:"primaryKey"`
	FirstName     string `grom:"size:100"`
	LastName      string `grom:"size:100"`
	MiddleName    string `grom:"size:100"`
	Email         string `gorm:"uniqueIndex"`
	Phone         string `gorm:"uniqueIndex"`
	Password      string `gorm:"size:255"`
	Verified      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Organisations []Organisation `gorm:"many2many:user_organisations;"`
}



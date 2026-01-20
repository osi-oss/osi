package models

type Permission struct {
	BaseModel

	Code        string `gorm:"not null;uniqueIndex"`
	Description string `gorm:"not null"`
	GroupName   string `gorm:"not null"`
}

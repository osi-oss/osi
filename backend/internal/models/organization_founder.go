package models

type OrganizationFounder struct {
	BaseModel

	OrganizationID int64 `gorm:"not null;index"`
	UserID         int64 `gorm:"not null;index"`

	SharePercent *float64
	IsMain       bool `gorm:"default:false"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	User         User         `gorm:"foreignKey:UserID"`
}

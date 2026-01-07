package models

type Location struct {
	BaseModel

	OrganizationID int64  `gorm:"not null;index"`
	Name           string `gorm:"not null"`
	Address        *string
	Source         string `gorm:"not null"` // 'registry' | 'manual'
	IsVerified     bool   `gorm:"not null;default:false"`
	IsActive       bool   `gorm:"default:true"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	Departments  []Department `gorm:"foreignKey:LocationID"`
}

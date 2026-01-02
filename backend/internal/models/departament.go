package models

type Department struct {
	BaseModel

	OrganizationID int64 `gorm:"not null;index"`
	ParentID       *int64

	Name        string `gorm:"not null"`
	Description *string

	Parent   *Department  `gorm:"foreignKey:ParentID"`
	Children []Department `gorm:"foreignKey:ParentID"`
}

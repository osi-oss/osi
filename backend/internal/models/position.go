package models

type Position struct {
	BaseModel

	OrganizationID int64  `gorm:"not null;index"`
	DepartmentID   *int64 `gorm:"index"`

	Name        string `gorm:"not null"`
	IsAdmin     bool   `gorm:"default:false"`
	Description *string

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	Department   *Department  `gorm:"foreignKey:DepartmentID"`
	Employees    []Employee   `gorm:"foreignKey:PositionID"`
}

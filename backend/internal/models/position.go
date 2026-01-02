package models

type Position struct {
	BaseModel

	OrganizationID int64 `gorm:"not null;index"`
	DepartmentID   *int64

	Name        string `gorm:"not null"`
	IsAdmin     bool
	Description *string
}

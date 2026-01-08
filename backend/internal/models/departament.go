package models

type Department struct {
	BaseModel

	LocationID int64  `gorm:"not null;index"`
	ParentID   *int64 `gorm:"index"`

	Name        string `gorm:"not null"`
	Description *string

	Location  Location     `gorm:"foreignKey:LocationID"`
	Parent    *Department  `gorm:"foreignKey:ParentID"`
	Children  []Department `gorm:"foreignKey:ParentID"`
	Positions []Position   `gorm:"foreignKey:DepartmentID"`
}

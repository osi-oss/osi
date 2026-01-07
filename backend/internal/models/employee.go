package models

import "time"

type Employee struct {
	BaseModel

	MemberID   int64 `gorm:"not null;index"`
	PositionID int64 `gorm:"not null;index"`

	IsIntern  bool `gorm:"default:false"`
	StartDate *time.Time
	EndDate   *time.Time

	Member   OrganizationMember `gorm:"foreignKey:MemberID"`
	Position Position           `gorm:"foreignKey:PositionID"`
}

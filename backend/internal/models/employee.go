package models

import "time"

type Employee struct {
	BaseModel

	OrganizationID int64        `gorm:"not null;index"`
	UserID         int64        `gorm:"not null;index"`
	PositionID     int64        `gorm:"not null;index"`
	Status         MemberStatus `gorm:"type:member_status;not null;default:'invited'"`

	IsIntern bool `gorm:"default:false"`

	StartDate *time.Time // When the employee accepted the invitation and started working
	EndDate   *time.Time // When the employee left the organization
	JoinedAt  *time.Time // When the employee accepted the invitation (alias for StartDate)

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	User         User         `gorm:"foreignKey:UserID"`
	Position     Position     `gorm:"foreignKey:PositionID"`
}

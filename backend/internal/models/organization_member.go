package models

import "time"

type OrganizationMember struct {
	BaseModel

	OrganizationID int64        `gorm:"not null;index"`
	UserID         int64        `gorm:"not null;index"`
	Status         MemberStatus `gorm:"type:member_status;not null;default:'invited'"`
	JoinedAt       *time.Time

	Employees []Employee
}

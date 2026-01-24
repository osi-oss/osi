package models

import "time"

type Invite struct {
	BaseModel

	OrganizationID  int64        `gorm:"not null;index"`
	PositionID      int64        `gorm:"not null;index"`
	InvitedUserID   *int64       `gorm:"index"` // NULLABLE if user not registered yet
	InvitedByUserID int64        `gorm:"not null;index"`
	Status          InviteStatus `gorm:"type:varchar(20);not null;default:'pending'"`

	InvitedEmail string     `gorm:"not null"`
	InvitedAt    *time.Time `gorm:"autoCreateTime:milli"`
	AcceptedAt   *time.Time
	DeclinedAt   *time.Time

	// Foreign keys
	Organization  Organization `gorm:"foreignKey:OrganizationID"`
	Position      Position     `gorm:"foreignKey:PositionID"`
	InvitedUser   *User        `gorm:"foreignKey:InvitedUserID"` // NULLABLE
	InvitedByUser User         `gorm:"foreignKey:InvitedByUserID;references:ID"`
}

// TableName specifies the table name
func (Invite) TableName() string {
	return "invites"
}

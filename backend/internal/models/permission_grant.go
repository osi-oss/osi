package models

import "time"

// PositionPermissionGrant - права должности с scope
// Все сотрудники на этой должности наследуют эти права
type PositionPermissionGrant struct {
	ID           int64     `gorm:"primaryKey"`
	PositionID   int64     `gorm:"not null;index"`
	PermissionID int64     `gorm:"not null;index"`
	ScopeType    ScopeType `gorm:"type:scope_type;not null;default:'organization'"`
	ScopeID      *int64    `gorm:"index"` // NULL = whole organization
	CreatedAt    time.Time

	Position   Position   `gorm:"foreignKey:PositionID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}

func (PositionPermissionGrant) TableName() string {
	return "position_permission_grants"
}

// EmployeePermissionGrant - индивидуальные права сотрудника с scope
// Дополняют права должности
type EmployeePermissionGrant struct {
	ID           int64     `gorm:"primaryKey"`
	EmployeeID   int64     `gorm:"not null;index"`
	PermissionID int64     `gorm:"not null;index"`
	ScopeType    ScopeType `gorm:"type:scope_type;not null;default:'organization'"`
	ScopeID      *int64    `gorm:"index"` // NULL = whole organization
	CreatedAt    time.Time

	Employee   Employee   `gorm:"foreignKey:EmployeeID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}

func (EmployeePermissionGrant) TableName() string {
	return "employee_permission_grants"
}

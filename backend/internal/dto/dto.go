package dto

import "time"

// ===== User DTOs =====

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserResponse struct {
	ID            int64     `json:"id"`
	Email         *string   `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// ===== Organization DTOs =====

type CreateOrganizationRequest struct {
	Name         string   `json:"name" binding:"required"`
	LegalName    *string  `json:"legal_name"`
	INN          *string  `json:"inn"`
	OGRN         *string  `json:"ogrn"`
	KPP          *string  `json:"kpp"`
	LegalAddress *string  `json:"legal_address"`
	SharePercent *float64 `json:"share_percent"`
}

type UpdateOrganizationRequest struct {
	Name         string  `json:"name"`
	LegalName    *string `json:"legal_name"`
	INN          *string `json:"inn"`
	OGRN         *string `json:"ogrn"`
	KPP          *string `json:"kpp"`
	LegalAddress *string `json:"legal_address"`
}

type OrganizationResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	LegalName    *string   `json:"legal_name"`
	INN          *string   `json:"inn"`
	OGRN         *string   `json:"ogrn"`
	KPP          *string   `json:"kpp"`
	LegalAddress *string   `json:"legal_address"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ===== Location DTOs =====

type CreateLocationRequest struct {
	Name       string  `json:"name" binding:"required"`
	Address    *string `json:"address"`
	Source     string  `json:"source" binding:"required"`
	IsVerified bool    `json:"is_verified"`
}

type UpdateLocationRequest struct {
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	Source     string  `json:"source"`
	IsVerified *bool   `json:"is_verified"`
	IsActive   *bool   `json:"is_active"`
}

type LocationResponse struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	Name           string    `json:"name"`
	Address        *string   `json:"address"`
	Source         string    `json:"source"`
	IsVerified     bool      `json:"is_verified"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ===== Department DTOs =====

type CreateDepartmentRequest struct {
	Name        string  `json:"name" binding:"required"`
	ParentID    *int64  `json:"parent_id"`
	Description *string `json:"description"`
}

type UpdateDepartmentRequest struct {
	Name        string  `json:"name"`
	ParentID    *int64  `json:"parent_id"`
	Description *string `json:"description"`
}

type DepartmentResponse struct {
	ID          int64     `json:"id"`
	LocationID  int64     `json:"location_id"`
	ParentID    *int64    `json:"parent_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ===== Position DTOs =====

type CreatePositionRequest struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name" binding:"required"`
	IsAdmin      bool    `json:"is_admin"`
	Description  *string `json:"description"`
}

type UpdatePositionRequest struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name"`
	IsAdmin      *bool   `json:"is_admin"`
	Description  *string `json:"description"`
}

type PositionResponse struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	DepartmentID   *int64    `json:"department_id"`
	Name           string    `json:"name"`
	IsAdmin        bool      `json:"is_admin"`
	Description    *string   `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ===== Member DTOs =====

type InviteMemberRequest struct {
	OrganizationID int64  `json:"organization_id" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
}

type MemberResponse struct {
	ID             int64        `json:"id"`
	OrganizationID int64        `json:"organization_id"`
	UserID         int64        `json:"user_id"`
	Status         string       `json:"status"`
	JoinedAt       *time.Time   `json:"joined_at"`
	User           UserResponse `json:"user,omitempty"`
}

// ===== Employee DTOs =====

type AssignPositionRequest struct {
	MemberID   int64      `json:"member_id" binding:"required"`
	PositionID int64      `json:"position_id" binding:"required"`
	StartDate  *time.Time `json:"start_date"`
	IsIntern   bool       `json:"is_intern"`
}

type UpdateEmployeeRequest struct {
	EndDate  *time.Time `json:"end_date"`
	IsIntern *bool      `json:"is_intern"`
}

type EmployeeResponse struct {
	ID         int64            `json:"id"`
	MemberID   int64            `json:"member_id"`
	PositionID int64            `json:"position_id"`
	IsIntern   bool             `json:"is_intern"`
	StartDate  *time.Time       `json:"start_date"`
	EndDate    *time.Time       `json:"end_date"`
	Position   PositionResponse `json:"position,omitempty"`
}

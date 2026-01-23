package models

// ===== User Status =====

type UserStatus string

const (
	UserStatusPendingEmail   UserStatus = "pending_email"   // Ожидание подтверждения email
	UserStatusPendingProfile UserStatus = "pending_profile" // Email подтверждён, профиль не заполнен
	UserStatusActive         UserStatus = "active"          // Полный доступ к системе
)

// IsValid проверяет валидность статуса пользователя
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPendingEmail, UserStatusPendingProfile, UserStatusActive:
		return true
	}
	return false
}

// ===== Organization Status =====

type OrgStatus string

const (
	OrgDraft    OrgStatus = "draft"
	OrgPending  OrgStatus = "pending"
	OrgApproved OrgStatus = "approved"
	OrgRejected OrgStatus = "rejected"
)

// ===== Member Status =====

type MemberStatus string

const (
	MemberInvited MemberStatus = "invited"
	MemberActive  MemberStatus = "active"
	MemberBlocked MemberStatus = "blocked"
)

// ===== Scope Type =====

type ScopeType string

const (
	ScopeOrganization ScopeType = "organization"
	ScopeLocation     ScopeType = "location"
	ScopeDepartment   ScopeType = "department"
)

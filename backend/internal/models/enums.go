package models

type OrgStatus string

const (
	OrgDraft    OrgStatus = "draft"
	OrgPending  OrgStatus = "pending"
	OrgApproved OrgStatus = "approved"
	OrgRejected OrgStatus = "rejected"
)

type MemberStatus string

const (
	MemberInvited MemberStatus = "invited"
	MemberActive  MemberStatus = "active"
	MemberBlocked MemberStatus = "blocked"
)

type ScopeType string

const (
	ScopeOrganization ScopeType = "organization"
	ScopeLocation     ScopeType = "location"
	ScopeDepartment   ScopeType = "department"
)

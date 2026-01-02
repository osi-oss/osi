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

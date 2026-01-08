package services

import (
	"errors"
	"time"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrMemberNotFound      = errors.New("member not found")
	ErrUserAlreadyMember   = errors.New("user is already a member of this organization")
	ErrCannotRemoveFounder = errors.New("cannot remove founder from members")
)

type MemberService struct {
	orgRepo       *repository.OrganizationRepository
	userRepo      *repository.UserRepository
	permissionSvc *PermissionService
}

func NewMemberService(
	orgRepo *repository.OrganizationRepository,
	userRepo *repository.UserRepository,
	permissionSvc *PermissionService,
) *MemberService {
	return &MemberService{
		orgRepo:       orgRepo,
		userRepo:      userRepo,
		permissionSvc: permissionSvc,
	}
}

type InviteMemberInput struct {
	OrganizationID int64  `json:"organization_id" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
}

// InviteMember invites a user to join an organization (requires members.invite permission)
func (s *MemberService) InviteMember(actorUserID int64, input InviteMemberInput) (*models.OrganizationMember, error) {
	// Check if actor has permission to invite members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, input.OrganizationID, "members.invite")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user with this email not found")
		}
		return nil, err
	}

	// Check if user is already a member
	existingMember, err := s.orgRepo.GetMemberByUserAndOrgID(user.ID, input.OrganizationID)
	if err == nil && existingMember != nil && existingMember.ID > 0 {
		return nil, ErrUserAlreadyMember
	}

	// Check if user is a founder
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(user.ID, input.OrganizationID)
	if err == nil && founder != nil && founder.ID > 0 {
		return nil, errors.New("user is already a founder of this organization")
	}

	// Create member with invited status
	member := &models.OrganizationMember{
		OrganizationID: input.OrganizationID,
		UserID:         user.ID,
		Status:         models.MemberInvited,
	}

	if err := s.orgRepo.CreateMember(member); err != nil {
		return nil, err
	}

	// Load user relation
	member.User = *user

	return member, nil
}

// AcceptInvitation changes member status from invited to active
func (s *MemberService) AcceptInvitation(userID int64, orgID int64) error {
	// Get the member
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	// Check if already active
	if member.Status == models.MemberActive {
		return errors.New("invitation already accepted")
	}

	// Update status to active
	now := time.Now()
	member.Status = models.MemberActive
	member.JoinedAt = &now

	return s.orgRepo.UpdateMemberStatus(member.ID, models.MemberActive)
}

// GetOrganizationMembers returns all members of an organization (requires members.view permission)
func (s *MemberService) GetOrganizationMembers(actorUserID int64, orgID int64) ([]models.OrganizationMember, error) {
	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, orgID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return s.orgRepo.GetMembersByOrganizationID(orgID)
}

// GetMemberByID returns a member by ID (requires members.view permission)
func (s *MemberService) GetMemberByID(actorUserID int64, memberID int64) (*models.OrganizationMember, error) {
	// Get the member to find the organization
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return member, nil
}

// RemoveMember removes a member from an organization (requires members.invite permission)
// Note: Cannot remove founders
func (s *MemberService) RemoveMember(actorUserID int64, memberID int64) error {
	// Get the member to find the organization
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return err
	}

	// Check if user is a founder (founders cannot be removed through this method)
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(member.UserID, member.OrganizationID)
	if err == nil && founder != nil && founder.ID > 0 {
		return ErrCannotRemoveFounder
	}

	// Check if actor has permission to manage members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "members.invite")
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrAccessDenied
	}

	return s.orgRepo.DeleteMember(memberID)
}

// GetMemberPermissions returns all permission codes for a member
func (s *MemberService) GetMemberPermissions(actorUserID int64, memberID int64) ([]string, error) {
	// Get the member to find the organization
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return s.permissionSvc.GetUserPermissions(member.UserID, member.OrganizationID)
}

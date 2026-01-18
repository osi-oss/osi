package services

import (
	"errors"
	"time"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
)

type MemberService struct {
	orgRepo  *repository.OrganizationRepository
	userRepo *repository.UserRepository
}

func NewMemberService(
	orgRepo *repository.OrganizationRepository,
	userRepo *repository.UserRepository,
) *MemberService {
	return &MemberService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

func (s *MemberService) InviteMember(input dto.InviteMemberRequest) (*models.OrganizationMember, error) {
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("user with this email not found")
		}
		return nil, err
	}

	existingMember, err := s.orgRepo.GetMemberByUserAndOrgID(user.ID, input.OrganizationID)
	if err == nil && existingMember != nil && existingMember.ID > 0 {
		return nil, apperrors.ErrUserAlreadyMember
	}

	founder, err := s.orgRepo.GetFounderByUserAndOrgID(user.ID, input.OrganizationID)
	if err == nil && founder != nil && founder.ID > 0 {
		return nil, apperrors.BadRequest("user is already a founder of this organization")
	}

	member := &models.OrganizationMember{
		OrganizationID: input.OrganizationID,
		UserID:         user.ID,
		Status:         models.MemberInvited,
	}

	if err := s.orgRepo.CreateMember(member); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create member")
	}

	member.User = *user
	return member, nil
}

func (s *MemberService) AcceptInvitation(userID int64, orgID int64) error {
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrMemberNotFound
		}
		return err
	}

	if member.Status != models.MemberInvited {
		return apperrors.BadRequest("invitation already processed")
	}

	now := time.Now()
	member.Status = models.MemberActive
	member.JoinedAt = &now

	return s.orgRepo.UpdateMemberStatus(member.ID, models.MemberActive)
}

func (s *MemberService) DeclineInvitation(userID int64, orgID int64) error {
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrMemberNotFound
		}
		return err
	}

	if member.Status != models.MemberInvited {
		return apperrors.BadRequest("invitation already processed")
	}

	return s.orgRepo.DeleteMember(member.ID)
}

func (s *MemberService) BlockMember(memberID int64) error {
	_, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return apperrors.ErrMemberNotFound
	}

	return s.orgRepo.UpdateMemberStatus(memberID, models.MemberBlocked)
}

func (s *MemberService) UnblockMember(memberID int64) error {
	_, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return apperrors.ErrMemberNotFound
	}

	return s.orgRepo.UpdateMemberStatus(memberID, models.MemberActive)
}

func (s *MemberService) RemoveMember(memberID int64) error {
	_, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return apperrors.ErrMemberNotFound
	}

	return s.orgRepo.DeleteMember(memberID)
}

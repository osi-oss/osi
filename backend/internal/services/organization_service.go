package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

func (s *OrganizationService) CreateOrganization(userID int64, input dto.CreateOrganizationRequest) (*models.Organization, error) {
	if input.Name == "" {
		return nil, apperrors.BadRequest("organization name is required")
	}

	org := &models.Organization{
		Name:         input.Name,
		LegalName:    input.LegalName,
		INN:          input.INN,
		OGRN:         input.OGRN,
		KPP:          input.KPP,
		LegalAddress: input.LegalAddress,
		Status:       models.OrgDraft,
	}

	if err := s.orgRepo.Create(org); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create organization")
	}

	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		SharePercent:   input.SharePercent,
		IsMain:         true,
	}

	if err := s.orgRepo.CreateFounder(founder); err != nil {
		s.orgRepo.Delete(org.ID)
		return nil, apperrors.Wrap(err, 500, "failed to create founder")
	}

	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) GetOrganization(orgID int64, userID int64) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, apperrors.ErrOrganizationNotFound
	}

	hasAccess := false
	for _, founder := range org.Founders {
		if founder.UserID == userID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		// Check members
		for _, member := range org.Members {
			if member.UserID == userID && member.Status == models.MemberActive {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return org, nil
}

func (s *OrganizationService) GetUserOrganizations(userID int64) ([]models.Organization, error) {
	return s.orgRepo.GetByUserID(userID)
}

// UserHasAccessToOrganization checks if user has access to organization
func (s *OrganizationService) UserHasAccessToOrganization(userID int64, orgID int64) (bool, error) {
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err == nil && member != nil && member.ID > 0 && member.Status == models.MemberActive {
		return true, nil
	}

	return false, nil
}

func (s *OrganizationService) UpdateOrganization(orgID int64, userID int64, input dto.UpdateOrganizationRequest) (*models.Organization, error) {
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		org.Name = input.Name
	}
	org.LegalName = input.LegalName
	org.INN = input.INN
	org.OGRN = input.OGRN
	org.KPP = input.KPP
	org.LegalAddress = input.LegalAddress

	if err := s.orgRepo.Update(org); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to update organization")
	}

	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) DeleteOrganization(orgID int64, userID int64) error {
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	isMainFounder := false
	for _, founder := range org.Founders {
		if founder.UserID == userID && founder.IsMain {
			isMainFounder = true
			break
		}
	}

	if !isMainFounder {
		return apperrors.ErrForbidden
	}

	return s.orgRepo.Delete(org.ID)
}

func (s *OrganizationService) GetMembers(orgID int64, userID int64) ([]models.OrganizationMember, error) {
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}
	return s.orgRepo.GetMembersByOrganizationID(orgID)
}

func (s *OrganizationService) UpdateMemberStatus(orgID int64, memberID int64, userID int64, status models.MemberStatus) error {
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}
	return s.orgRepo.UpdateMemberStatus(memberID, status)
}

func (s *OrganizationService) RemoveMember(orgID int64, memberID int64, userID int64) error {
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}
	return s.orgRepo.DeleteMember(memberID)
}

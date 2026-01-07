package services

import (
	"errors"
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrUnauthorized         = errors.New("unauthorized")
)

type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

type CreateOrganizationInput struct {
	Name         string   `json:"name" binding:"required"`
	LegalName    *string  `json:"legal_name"`
	INN          *string  `json:"inn"`
	OGRN         *string  `json:"ogrn"`
	KPP          *string  `json:"kpp"`
	LegalAddress *string  `json:"legal_address"`
	SharePercent *float64 `json:"share_percent"`
}

func (s *OrganizationService) CreateOrganization(userID int64, input CreateOrganizationInput) (*models.Organization, error) {
	// Валидация входных данных
	if input.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Создаем организацию
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
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Создаем основателя (founder) - связываем пользователя с организацией
	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		SharePercent:   input.SharePercent,
		IsMain:         true, // Создатель организации - главный основатель
	}

	if err := s.orgRepo.CreateFounder(founder); err != nil {
		// Если не удалось создать основателя, удаляем организацию
		s.orgRepo.Delete(org.ID)
		return nil, fmt.Errorf("failed to create founder: %w", err)
	}

	// Загружаем организацию с основателями для возврата
	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) GetOrganization(orgID int64, userID int64) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	// Проверяем, что пользователь имеет доступ к этой организации
	hasAccess := false
	for _, founder := range org.Founders {
		if founder.UserID == userID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return org, nil
}

func (s *OrganizationService) GetUserOrganizations(userID int64) ([]models.Organization, error) {
	return s.orgRepo.GetByUserID(userID)
}

func (s *OrganizationService) UpdateOrganization(orgID int64, userID int64, input CreateOrganizationInput) (*models.Organization, error) {
	// Проверяем доступ
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if input.Name != "" {
		org.Name = input.Name
	}
	org.LegalName = input.LegalName
	org.INN = input.INN
	org.OGRN = input.OGRN
	org.KPP = input.KPP
	org.LegalAddress = input.LegalAddress

	if err := s.orgRepo.Update(org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) DeleteOrganization(orgID int64, userID int64) error {
	// Проверяем доступ
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	// Проверяем, что пользователь - главный основатель
	isMainFounder := false
	for _, founder := range org.Founders {
		if founder.UserID == userID && founder.IsMain {
			isMainFounder = true
			break
		}
	}

	if !isMainFounder {
		return fmt.Errorf("only main founder can delete organization")
	}

	return s.orgRepo.Delete(org.ID)
}

// Member methods
// type InviteMemberInput struct {
// 	UserID int64 `json:"user_id" binding:"required"`
// }

// func (s *OrganizationService) InviteMember(orgID int64, userID int64, input InviteMemberInput) (*models.OrganizationMember, error) {
// 	// Проверяем доступ
// 	_, err := s.GetOrganization(orgID, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	member := &models.OrganizationMember{
// 		OrganizationID: orgID,
// 		UserID:         input.UserID,
// 		Status:         models.MemberInvited,
// 	}

// 	if err := s.orgRepo.CreateMember(member); err != nil {
// 		return nil, fmt.Errorf("failed to invite member: %w", err)
// 	}

// 	return s.orgRepo.GetMemberByID(member.ID)
// }

func (s *OrganizationService) GetMembers(orgID int64, userID int64) ([]models.OrganizationMember, error) {
	// Проверяем доступ
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	return s.orgRepo.GetMembersByOrganizationID(orgID)
}

func (s *OrganizationService) UpdateMemberStatus(orgID int64, memberID int64, userID int64, status models.MemberStatus) error {
	// Проверяем доступ
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	return s.orgRepo.UpdateMemberStatus(memberID, status)
}

func (s *OrganizationService) RemoveMember(orgID int64, memberID int64, userID int64) error {
	// Проверяем доступ
	_, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	return s.orgRepo.DeleteMember(memberID)
}

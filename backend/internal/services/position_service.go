package services

import (
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type PositionService struct {
	posRepo    *repository.PositionRepository
	orgService *OrganizationService
}

func NewPositionService(posRepo *repository.PositionRepository, orgService *OrganizationService) *PositionService {
	return &PositionService{
		posRepo:    posRepo,
		orgService: orgService,
	}
}

type CreatePositionInput struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name" binding:"required"`
	IsAdmin      bool    `json:"is_admin"`
	Description  *string `json:"description"`
}

func (s *PositionService) CreatePosition(orgID int64, userID int64, input CreatePositionInput) (*models.Position, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	position := &models.Position{
		OrganizationID: orgID,
		DepartmentID:   input.DepartmentID,
		Name:           input.Name,
		IsAdmin:        input.IsAdmin,
		Description:    input.Description,
	}

	if err := s.posRepo.Create(position); err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	return s.posRepo.GetByID(position.ID)
}

func (s *PositionService) GetPosition(posID int64, orgID int64, userID int64) (*models.Position, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	return s.posRepo.GetByID(posID)
}

// func (s *PositionService) GetPositionsByOrganization(orgID int64, userID int64) ([]models.Position, error) {
// 	// Проверяем доступ к организации
// 	_, err := s.orgService.GetOrganization(orgID, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return s.posRepo.GetByOrganizationID(orgID)
// }

// func (s *PositionService) GetPositionsByDepartment(deptID int64, orgID int64, userID int64) ([]models.Position, error) {
// 	// Проверяем доступ к организации
// 	_, err := s.orgService.GetOrganization(orgID, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return s.posRepo.GetByDepartmentID(deptID)
// }

func (s *PositionService) UpdatePosition(posID int64, orgID int64, userID int64, input CreatePositionInput) (*models.Position, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, fmt.Errorf("position not found: %w", err)
	}

	pos.Name = input.Name
	pos.DepartmentID = input.DepartmentID
	pos.IsAdmin = input.IsAdmin
	pos.Description = input.Description

	if err := s.posRepo.Update(pos); err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return s.posRepo.GetByID(pos.ID)
}

func (s *PositionService) DeletePosition(posID int64, orgID int64, userID int64) error {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	return s.posRepo.Delete(posID)
}

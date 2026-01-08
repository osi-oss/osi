package services

import (
	"errors"
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

var (
	ErrPositionNotFound = errors.New("position not found")
)

type PositionService struct {
	posRepo       *repository.PositionRepository
	orgService    *OrganizationService
	permissionSvc *PermissionService
}

func NewPositionService(posRepo *repository.PositionRepository, orgService *OrganizationService, permissionSvc *PermissionService) *PositionService {
	return &PositionService{
		posRepo:       posRepo,
		orgService:    orgService,
		permissionSvc: permissionSvc,
	}
}

type CreatePositionInput struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name" binding:"required"`
	IsAdmin      bool    `json:"is_admin"`
	Description  *string `json:"description"`
}

type UpdatePositionInput struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name"`
	IsAdmin      *bool   `json:"is_admin"`
	Description  *string `json:"description"`
}

func (s *PositionService) CreatePosition(orgID int64, userID int64, input CreatePositionInput) (*models.Position, error) {
	// Check if user has permission to create positions
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, orgID, "positions.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
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

// GetPosition получает позицию по ID
func (s *PositionService) GetPosition(posID int64, userID int64) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, ErrPositionNotFound
	}

	// Check if user has access to this organization
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, pos.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return pos, nil
}

// GetOrganizationPositions получает все позиции организации
func (s *PositionService) GetOrganizationPositions(orgID int64, userID int64) ([]models.Position, error) {
	// Check if user has access to this organization
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.posRepo.GetByOrganizationID(orgID)
}

// GetDepartmentPositions получает все позиции отдела
func (s *PositionService) GetDepartmentPositions(deptID int64, userID int64) ([]models.Position, error) {
	positions, err := s.posRepo.GetByDepartmentID(deptID)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return positions, nil
	}

	// Check if user has access to the organization of these positions
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, positions[0].OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return positions, nil
}

// UpdatePosition обновляет позицию
func (s *PositionService) UpdatePosition(posID int64, userID int64, input UpdatePositionInput) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, ErrPositionNotFound
	}

	// Check if user has permission to update positions
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, pos.OrganizationID, "positions.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	// Обновляем только переданные поля
	if input.Name != "" {
		pos.Name = input.Name
	}
	if input.DepartmentID != nil {
		pos.DepartmentID = input.DepartmentID
	}
	if input.IsAdmin != nil {
		pos.IsAdmin = *input.IsAdmin
	}
	if input.Description != nil {
		pos.Description = input.Description
	}

	if err := s.posRepo.Update(pos); err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return s.posRepo.GetByID(pos.ID)
}

// DeletePosition удаляет позицию
func (s *PositionService) DeletePosition(posID int64, userID int64) error {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return ErrPositionNotFound
	}

	// Check if user has permission to delete positions
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, pos.OrganizationID, "positions.create")
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrAccessDenied
	}

	return s.posRepo.Delete(posID)
}

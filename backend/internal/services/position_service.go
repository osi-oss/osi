package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
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

func (s *PositionService) CreatePosition(orgID int64, userID int64, input dto.CreatePositionRequest) (*models.Position, error) {
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, orgID, "positions.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
	}

	position := &models.Position{
		OrganizationID: orgID,
		DepartmentID:   input.DepartmentID,
		Name:           input.Name,
		IsAdmin:        input.IsAdmin,
		Description:    input.Description,
	}

	if err := s.posRepo.Create(position); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create position")
	}

	return s.posRepo.GetByID(position.ID)
}

func (s *PositionService) GetPosition(posID int64, userID int64) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, apperrors.ErrPositionNotFound
	}

	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, pos.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return pos, nil
}

func (s *PositionService) GetOrganizationPositions(orgID int64, userID int64) ([]models.Position, error) {
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return s.posRepo.GetByOrganizationID(orgID)
}

func (s *PositionService) GetDepartmentPositions(deptID int64, userID int64) ([]models.Position, error) {
	positions, err := s.posRepo.GetByDepartmentID(deptID)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return positions, nil
	}

	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, positions[0].OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return positions, nil
}

func (s *PositionService) UpdatePosition(posID int64, userID int64, input dto.UpdatePositionRequest) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, apperrors.ErrPositionNotFound
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, pos.OrganizationID, "positions.update")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
	}

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
		return nil, apperrors.Wrap(err, 500, "failed to update position")
	}

	return s.posRepo.GetByID(pos.ID)
}

func (s *PositionService) DeletePosition(posID int64, userID int64) error {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return apperrors.ErrPositionNotFound
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, pos.OrganizationID, "positions.delete")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	return s.posRepo.Delete(posID)
}

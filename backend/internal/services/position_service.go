package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type PositionService struct {
	posRepo *repository.PositionRepository
}

func NewPositionService(posRepo *repository.PositionRepository) *PositionService {
	return &PositionService{
		posRepo: posRepo,
	}
}

func (s *PositionService) CreatePosition(orgID int64, input dto.CreatePositionRequest) (*models.Position, error) {
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

func (s *PositionService) GetPosition(posID int64) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, apperrors.ErrPositionNotFound
	}
	return pos, nil
}

func (s *PositionService) GetOrganizationPositions(orgID int64) ([]models.Position, error) {
	return s.posRepo.GetByOrganizationID(orgID)
}

func (s *PositionService) GetDepartmentPositions(deptID int64) ([]models.Position, error) {
	return s.posRepo.GetByDepartmentID(deptID)
}

func (s *PositionService) UpdatePosition(posID int64, input dto.UpdatePositionRequest) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, apperrors.ErrPositionNotFound
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

func (s *PositionService) DeletePosition(posID int64) error {
	_, err := s.posRepo.GetByID(posID)
	if err != nil {
		return apperrors.ErrPositionNotFound
	}

	return s.posRepo.Delete(posID)
}

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

type UpdatePositionInput struct {
	DepartmentID *int64  `json:"department_id"`
	Name         string  `json:"name"`
	IsAdmin      *bool   `json:"is_admin"`
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

// GetPosition получает позицию по ID
func (s *PositionService) GetPosition(posID int64, userID int64) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, ErrPositionNotFound
	}

	// Проверяем доступ к организации этой позиции
	_, err = s.orgService.GetOrganization(pos.OrganizationID, userID)
	if err != nil {
		return nil, err
	}

	return pos, nil
}

// GetOrganizationPositions получает все позиции организации
func (s *PositionService) GetOrganizationPositions(orgID int64, userID int64) ([]models.Position, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
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
		// Если позиций нет, нужно проверить существование отдела через другой способ
		// Для простоты просто возвращаем пустой массив
		return positions, nil
	}

	// Проверяем доступ к организации первой позиции (все позиции отдела принадлежат одной организации)
	_, err = s.orgService.GetOrganization(positions[0].OrganizationID, userID)
	if err != nil {
		return nil, err
	}

	return positions, nil
}

// UpdatePosition обновляет позицию
func (s *PositionService) UpdatePosition(posID int64, userID int64, input UpdatePositionInput) (*models.Position, error) {
	pos, err := s.posRepo.GetByID(posID)
	if err != nil {
		return nil, ErrPositionNotFound
	}

	// Проверяем доступ к организации этой позиции
	_, err = s.orgService.GetOrganization(pos.OrganizationID, userID)
	if err != nil {
		return nil, err
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

	// Проверяем доступ к организации этой позиции
	_, err = s.orgService.GetOrganization(pos.OrganizationID, userID)
	if err != nil {
		return err
	}

	return s.posRepo.Delete(posID)
}

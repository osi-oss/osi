package services

import (
	"errors"
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

var (
	ErrDepartmentNotFound = errors.New("department not found")
)

type DepartmentService struct {
	deptRepo        *repository.DepartmentRepository
	locationService *LocationService
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository, locationService *LocationService) *DepartmentService {
	return &DepartmentService{
		deptRepo:        deptRepo,
		locationService: locationService,
	}
}

type CreateDepartmentInput struct {
	Name        string  `json:"name" binding:"required"`
	ParentID    *int64  `json:"parent_id"`
	Description *string `json:"description"`
}

type UpdateDepartmentInput struct {
	Name        string  `json:"name"`
	ParentID    *int64  `json:"parent_id"`
	Description *string `json:"description"`
}

// CreateDepartment создает новый отдел в локации
func (s *DepartmentService) CreateDepartment(locationID int64, userID int64, input CreateDepartmentInput) (*models.Department, error) {
	// Проверяем доступ к локации
	location, err := s.locationService.GetLocation(locationID, userID)
	if err != nil {
		return nil, err
	}

	// Если указан родительский отдел, проверяем что он существует и принадлежит той же локации
	if input.ParentID != nil {
		parent, err := s.deptRepo.GetByID(*input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent department not found: %w", err)
		}
		if parent.LocationID != locationID {
			return nil, fmt.Errorf("parent department belongs to different location")
		}
	}

	department := &models.Department{
		LocationID:  location.ID,
		ParentID:    input.ParentID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.deptRepo.Create(department); err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return s.deptRepo.GetByID(department.ID)
}

// GetDepartment получает отдел по ID
func (s *DepartmentService) GetDepartment(deptID int64, userID int64) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}

	// Проверяем доступ к локации этого отдела
	_, err = s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return nil, err
	}

	return dept, nil
}

// GetLocationDepartments получает все отделы локации
func (s *DepartmentService) GetLocationDepartments(locationID int64, userID int64) ([]models.Department, error) {
	// Проверяем доступ к локации
	_, err := s.locationService.GetLocation(locationID, userID)
	if err != nil {
		return nil, err
	}

	return s.deptRepo.GetByLocationID(locationID)
}

// UpdateDepartment обновляет отдел
func (s *DepartmentService) UpdateDepartment(deptID int64, userID int64, input UpdateDepartmentInput) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, ErrDepartmentNotFound
	}

	// Проверяем доступ к локации этого отдела
	_, err = s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return nil, err
	}

	// Если указан новый родительский отдел, проверяем его
	if input.ParentID != nil {
		// Проверяем, что отдел не устанавливается родителем самому себе
		if *input.ParentID == deptID {
			return nil, fmt.Errorf("department cannot be its own parent")
		}

		parent, err := s.deptRepo.GetByID(*input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent department not found: %w", err)
		}
		if parent.LocationID != dept.LocationID {
			return nil, fmt.Errorf("parent department belongs to different location")
		}
	}

	// Обновляем только переданные поля
	if input.Name != "" {
		dept.Name = input.Name
	}
	if input.ParentID != nil {
		dept.ParentID = input.ParentID
	}
	if input.Description != nil {
		dept.Description = input.Description
	}

	if err := s.deptRepo.Update(dept); err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return s.deptRepo.GetByID(dept.ID)
}

// DeleteDepartment удаляет отдел
func (s *DepartmentService) DeleteDepartment(deptID int64, userID int64) error {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return ErrDepartmentNotFound
	}

	// Проверяем доступ к локации этого отдела
	_, err = s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return err
	}

	// Проверяем, что у отдела нет дочерних отделов
	if len(dept.Children) > 0 {
		return fmt.Errorf("cannot delete department with child departments")
	}

	return s.deptRepo.Delete(deptID)
}

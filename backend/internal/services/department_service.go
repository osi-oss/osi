package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type DepartmentService struct {
	deptRepo     *repository.DepartmentRepository
	locationRepo *repository.LocationRepository
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository, locationRepo *repository.LocationRepository) *DepartmentService {
	return &DepartmentService{
		deptRepo:     deptRepo,
		locationRepo: locationRepo,
	}
}

func (s *DepartmentService) CreateDepartment(locationID int64, input dto.CreateDepartmentRequest) (*models.Department, error) {
	// Check if location exists
	_, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, apperrors.ErrLocationNotFound
	}

	if input.ParentID != nil {
		parent, err := s.deptRepo.GetByID(*input.ParentID)
		if err != nil {
			return nil, apperrors.BadRequest("parent department not found")
		}
		if parent.LocationID != locationID {
			return nil, apperrors.BadRequest("parent department belongs to different location")
		}
	}

	department := &models.Department{
		LocationID:  locationID,
		ParentID:    input.ParentID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.deptRepo.Create(department); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create department")
	}

	return s.deptRepo.GetByID(department.ID)
}

func (s *DepartmentService) GetDepartment(deptID int64) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, apperrors.ErrDepartmentNotFound
	}
	return dept, nil
}

func (s *DepartmentService) GetLocationDepartments(locationID int64) ([]models.Department, error) {
	return s.deptRepo.GetByLocationID(locationID)
}

func (s *DepartmentService) UpdateDepartment(deptID int64, input dto.UpdateDepartmentRequest) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, apperrors.ErrDepartmentNotFound
	}

	if input.ParentID != nil {
		if *input.ParentID == deptID {
			return nil, apperrors.BadRequest("department cannot be its own parent")
		}
		parent, err := s.deptRepo.GetByID(*input.ParentID)
		if err != nil {
			return nil, apperrors.BadRequest("parent department not found")
		}
		if parent.LocationID != dept.LocationID {
			return nil, apperrors.BadRequest("parent department belongs to different location")
		}
	}

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
		return nil, apperrors.Wrap(err, 500, "failed to update department")
	}

	return s.deptRepo.GetByID(dept.ID)
}

func (s *DepartmentService) DeleteDepartment(deptID int64) error {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return apperrors.ErrDepartmentNotFound
	}

	if len(dept.Children) > 0 {
		return apperrors.BadRequest("cannot delete department with child departments")
	}

	return s.deptRepo.Delete(deptID)
}

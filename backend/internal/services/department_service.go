package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type DepartmentService struct {
	deptRepo        *repository.DepartmentRepository
	locationService *LocationService
	permissionSvc   *PermissionService
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository, locationService *LocationService, permissionSvc *PermissionService) *DepartmentService {
	return &DepartmentService{
		deptRepo:        deptRepo,
		locationService: locationService,
		permissionSvc:   permissionSvc,
	}
}

func (s *DepartmentService) CreateDepartment(locationID int64, userID int64, input dto.CreateDepartmentRequest) (*models.Department, error) {
	location, err := s.locationService.GetLocation(locationID, userID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "departments.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
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
		LocationID:  location.ID,
		ParentID:    input.ParentID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.deptRepo.Create(department); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create department")
	}

	return s.deptRepo.GetByID(department.ID)
}

func (s *DepartmentService) GetDepartment(deptID int64, userID int64) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, apperrors.ErrDepartmentNotFound
	}

	_, err = s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *DepartmentService) GetLocationDepartments(locationID int64, userID int64) ([]models.Department, error) {
	_, err := s.locationService.GetLocation(locationID, userID)
	if err != nil {
		return nil, err
	}

	return s.deptRepo.GetByLocationID(locationID)
}

func (s *DepartmentService) UpdateDepartment(deptID int64, userID int64, input dto.UpdateDepartmentRequest) (*models.Department, error) {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return nil, apperrors.ErrDepartmentNotFound
	}

	location, err := s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "departments.update")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
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

func (s *DepartmentService) DeleteDepartment(deptID int64, userID int64) error {
	dept, err := s.deptRepo.GetByID(deptID)
	if err != nil {
		return apperrors.ErrDepartmentNotFound
	}

	location, err := s.locationService.GetLocation(dept.LocationID, userID)
	if err != nil {
		return err
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "departments.delete")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	if len(dept.Children) > 0 {
		return apperrors.BadRequest("cannot delete department with child departments")
	}

	return s.deptRepo.Delete(deptID)
}

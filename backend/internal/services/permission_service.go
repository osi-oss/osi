package services

import (
	"errors"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
)

type PermissionService struct {
	permissionRepo      *repository.PermissionRepository
	permissionGrantRepo *repository.PermissionGrantRepository
	orgRepo             *repository.OrganizationRepository
	employeeRepo        *repository.EmployeeRepository
}

func NewPermissionService(
	permissionRepo *repository.PermissionRepository,
	permissionGrantRepo *repository.PermissionGrantRepository,
	orgRepo *repository.OrganizationRepository,
	employeeRepo *repository.EmployeeRepository,
) *PermissionService {
	return &PermissionService{
		permissionRepo:      permissionRepo,
		permissionGrantRepo: permissionGrantRepo,
		orgRepo:             orgRepo,
		employeeRepo:        employeeRepo,
	}
}

// UserHasAccessToOrganization checks if user has access to organization (founder or active member)
func (s *PermissionService) UserHasAccessToOrganization(userID int64, orgID int64) (bool, error) {
	// Check if user is a founder
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	// Check if user is an active member (has at least one active employee record)
	employee, err := s.orgRepo.GetEmployeeByUserAndOrgID(userID, orgID)
	if err == nil && employee != nil && employee.ID > 0 && employee.Status == models.MemberActive {
		return true, nil
	}

	return false, nil
}


// UserHasScopedPermissionWithHierarchy checks permission with full hierarchy support
// It validates that the scope matches the context hierarchy:
// - scope=organization: always applies
// - scope=location: applies if locationID matches in context
// - scope=department: applies if departmentID matches in context (and location matches)
func (s *PermissionService) UserHasScopedPermissionWithHierarchy(
	userID, orgID int64,
	permissionCode string,
	scopeType models.ScopeType,
	context *models.PermissionContext,
) (bool, error) {
	// Check if user is a founder (founders have all permissions)
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	// Get all employees for this user in this organization
	employees, err := s.orgRepo.GetEmployeesByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if len(employees) == 0 {
		return false, nil
	}

	// Collect active employee IDs and position IDs
	var employeeIDs []int64
	var positionIDs []int64
	for _, emp := range employees {
		if emp.Status == models.MemberActive && emp.EndDate == nil { // Active employment
			employeeIDs = append(employeeIDs, emp.ID)
			positionIDs = append(positionIDs, emp.PositionID)
		}
	}

	if len(employeeIDs) == 0 {
		return false, nil
	}

	// Check position grants with hierarchy
	hasPermission, err := s.permissionGrantRepo.CheckPositionHasScopedPermissionWithHierarchy(
		positionIDs,
		permissionCode,
		scopeType,
		context,
	)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Check employee grants with hierarchy
	hasPermission, err = s.permissionGrantRepo.CheckEmployeeHasScopedPermissionWithHierarchy(
		employeeIDs,
		permissionCode,
		scopeType,
		context,
	)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// === Scoped Permission Grant Management ===

// GrantPermissionToPosition grants a scoped permission to a position
func (s *PermissionService) GrantPermissionToPosition(positionID int64, permissionCode string, scopeType models.ScopeType, scopeID *int64) error {
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrPermissionNotFound
		}
		return err
	}

	grant := &models.PositionPermissionGrant{
		PositionID:   positionID,
		PermissionID: permission.ID,
		ScopeType:    scopeType,
		ScopeID:      scopeID,
	}

	return s.permissionGrantRepo.GrantToPosition(grant)
}

// GrantPermissionToEmployee grants a scoped permission to an employee
func (s *PermissionService) GrantPermissionToEmployee(employeeID int64, permissionCode string, scopeType models.ScopeType, scopeID *int64) error {
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrPermissionNotFound
		}
		return err
	}

	grant := &models.EmployeePermissionGrant{
		EmployeeID:   employeeID,
		PermissionID: permission.ID,
		ScopeType:    scopeType,
		ScopeID:      scopeID,
	}

	return s.permissionGrantRepo.GrantToEmployee(grant)
}

// RevokePermissionFromEmployee revokes a scoped permission from an employee
func (s *PermissionService) RevokePermissionFromEmployee(employeeID int64, permissionCode string, scopeType models.ScopeType, scopeID *int64) error {
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrPermissionNotFound
		}
		return err
	}

	return s.permissionGrantRepo.RevokeFromEmployee(employeeID, permission.ID, scopeType, scopeID)
}

// GetEmployeePermissionGrants returns all scoped permission grants for an employee
func (s *PermissionService) GetEmployeePermissionGrants(employeeID int64) ([]models.EmployeePermissionGrant, error) {
	return s.permissionGrantRepo.GetEmployeeGrants(employeeID)
}

// GetAllPermissions returns all available permissions
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.GetAll()
}

// GetPermissionByCode returns a permission by its code
func (s *PermissionService) GetPermissionByCode(code string) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrPermissionNotFound
		}
		return nil, err
	}
	return permission, nil
}

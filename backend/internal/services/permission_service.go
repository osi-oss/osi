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

	// Check if user is an active member
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err == nil && member != nil && member.ID > 0 && member.Status == models.MemberActive {
		return true, nil
	}

	return false, nil
}

// UserHasScopedPermission checks if a user has a specific scoped permission
// It follows this hierarchy:
// 1. If user is a founder -> has all permissions
// 2. Check position_permission_grants for all user's positions
// 3. Check employee_permission_grants for all user's employees
func (s *PermissionService) UserHasScopedPermission(userID, orgID int64, permissionCode string, scopeType models.ScopeType, scopeID *int64) (bool, error) {
	// Check if user is a founder (founders have all permissions)
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	// Get member
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if member.Status != models.MemberActive {
		return false, nil
	}

	// Collect active employee IDs and position IDs
	var employeeIDs []int64
	var positionIDs []int64
	for _, emp := range member.Employees {
		if emp.EndDate == nil { // Active employment
			employeeIDs = append(employeeIDs, emp.ID)
			positionIDs = append(positionIDs, emp.PositionID)
		}
	}

	// Check position grants
	hasPermission, err := s.permissionGrantRepo.CheckPositionHasScopedPermission(positionIDs, permissionCode, scopeType, scopeID)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Check employee grants
	hasPermission, err = s.permissionGrantRepo.CheckEmployeeHasScopedPermission(employeeIDs, permissionCode, scopeType, scopeID)
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

// RevokePermissionFromPosition revokes a scoped permission from a position
func (s *PermissionService) RevokePermissionFromPosition(positionID int64, permissionCode string, scopeType models.ScopeType, scopeID *int64) error {
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrPermissionNotFound
		}
		return err
	}

	return s.permissionGrantRepo.RevokeFromPosition(positionID, permission.ID, scopeType, scopeID)
}

// GetPositionPermissionGrants returns all scoped permission grants for a position
func (s *PermissionService) GetPositionPermissionGrants(positionID int64) ([]models.PositionPermissionGrant, error) {
	return s.permissionGrantRepo.GetPositionGrants(positionID)
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

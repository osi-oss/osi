package services

import (
	"errors"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
)

type PermissionService struct {
	permissionRepo *repository.PermissionRepository
	orgRepo        *repository.OrganizationRepository
	employeeRepo   *repository.EmployeeRepository
}

func NewPermissionService(
	permissionRepo *repository.PermissionRepository,
	orgRepo *repository.OrganizationRepository,
	employeeRepo *repository.EmployeeRepository,
) *PermissionService {
	return &PermissionService{
		permissionRepo: permissionRepo,
		orgRepo:        orgRepo,
		employeeRepo:   employeeRepo,
	}
}

// UserHasPermission checks if a user has a specific permission in an organization
// It follows this hierarchy:
// 1. If user is a founder -> has all permissions
// 2. If user is a member with employees (has positions) -> check member permissions + all position permissions
// 3. If user is a member without positions -> check only member permissions
func (s *PermissionService) UserHasPermission(userID int64, orgID int64, permissionCode string) (bool, error) {
	// Check if user is a founder (founders have all permissions)
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	// Check if user is a member
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // User is not a member of this organization
		}
		return false, err
	}

	// Check if member is active
	if member.Status != models.MemberActive {
		return false, nil
	}

	// Get the permission to check
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrPermissionNotFound
		}
		return false, err
	}

	// Check member's direct permissions
	for _, perm := range member.Permissions {
		if perm.ID == permission.ID {
			return true, nil
		}
	}

	// Check permissions from all positions this member holds
	for _, employee := range member.Employees {
		// Skip if employment has ended
		if employee.EndDate != nil {
			continue
		}

		// Check if this position has the permission
		for _, perm := range employee.Position.Permissions {
			if perm.ID == permission.ID {
				return true, nil
			}
		}
	}

	return false, nil
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

// UserHasAnyPermission checks if a user has any of the specified permissions
func (s *PermissionService) UserHasAnyPermission(userID int64, orgID int64, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		hasPermission, err := s.UserHasPermission(userID, orgID, code)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

// UserHasAllPermissions checks if a user has all of the specified permissions
func (s *PermissionService) UserHasAllPermissions(userID int64, orgID int64, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		hasPermission, err := s.UserHasPermission(userID, orgID, code)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}
	return true, nil
}

// GetUserPermissions returns all permission codes for a user in an organization
func (s *PermissionService) GetUserPermissions(userID int64, orgID int64) ([]string, error) {
	// Check if user is a founder (founders have all permissions)
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		// Return all permissions
		allPermissions, err := s.permissionRepo.GetAll()
		if err != nil {
			return nil, err
		}
		codes := make([]string, len(allPermissions))
		for i, perm := range allPermissions {
			codes[i] = perm.Code
		}
		return codes, nil
	}

	// Check if user is a member
	member, err := s.orgRepo.GetMemberByUserAndOrgID(userID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, err
	}

	// Check if member is active
	if member.Status != models.MemberActive {
		return []string{}, nil
	}

	// Collect unique permissions
	permissionMap := make(map[string]bool)

	// Add member's direct permissions
	for _, perm := range member.Permissions {
		permissionMap[perm.Code] = true
	}

	// Add permissions from all positions
	for _, employee := range member.Employees {
		// Skip if employment has ended
		if employee.EndDate != nil {
			continue
		}

		for _, perm := range employee.Position.Permissions {
			permissionMap[perm.Code] = true
		}
	}

	// Convert map to slice
	codes := make([]string, 0, len(permissionMap))
	for code := range permissionMap {
		codes = append(codes, code)
	}

	return codes, nil
}

// AssignPermissionToMember assigns a permission to a member (requires permissions.manage)
func (s *PermissionService) AssignPermissionToMember(actorUserID int64, memberID int64, permissionCode string) error {
	// Get the member to find the organization
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return err
	}

	// Check if actor has permission to manage permissions
	hasPermission, err := s.UserHasPermission(actorUserID, member.OrganizationID, "permissions.manage")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	// Get the permission
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		return err
	}

	return s.permissionRepo.AssignPermissionToMember(memberID, permission.ID)
}

// RemovePermissionFromMember removes a permission from a member (requires permissions.manage)
func (s *PermissionService) RemovePermissionFromMember(actorUserID int64, memberID int64, permissionCode string) error {
	// Get the member to find the organization
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return err
	}

	// Check if actor has permission to manage permissions
	hasPermission, err := s.UserHasPermission(actorUserID, member.OrganizationID, "permissions.manage")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	// Get the permission
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		return err
	}

	return s.permissionRepo.RemovePermissionFromMember(memberID, permission.ID)
}

// AssignPermissionToPosition assigns a permission to a position (requires permissions.manage)
func (s *PermissionService) AssignPermissionToPosition(actorUserID int64, positionID int64, orgID int64, permissionCode string) error {
	// Check if actor has permission to manage permissions
	hasPermission, err := s.UserHasPermission(actorUserID, orgID, "permissions.manage")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	// Get the permission
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		return err
	}

	return s.permissionRepo.AssignPermissionToPosition(positionID, permission.ID)
}

// RemovePermissionFromPosition removes a permission from a position (requires permissions.manage)
func (s *PermissionService) RemovePermissionFromPosition(actorUserID int64, positionID int64, orgID int64, permissionCode string) error {
	// Check if actor has permission to manage permissions
	hasPermission, err := s.UserHasPermission(actorUserID, orgID, "permissions.manage")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	// Get the permission
	permission, err := s.permissionRepo.GetByCode(permissionCode)
	if err != nil {
		return err
	}

	return s.permissionRepo.RemovePermissionFromPosition(positionID, permission.ID)
}

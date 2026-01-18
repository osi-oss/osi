package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type PermissionGrantRepository struct {
	db *gorm.DB
}

func NewPermissionGrantRepository(db *gorm.DB) *PermissionGrantRepository {
	return &PermissionGrantRepository{db: db}
}

// === Position Permission Grants ===

// GrantToPosition adds a scoped permission to a position
func (r *PermissionGrantRepository) GrantToPosition(grant *models.PositionPermissionGrant) error {
	return r.db.Create(grant).Error
}

// RevokeFromPosition removes a scoped permission from a position
func (r *PermissionGrantRepository) RevokeFromPosition(positionID, permissionID int64, scopeType models.ScopeType, scopeID *int64) error {
	query := r.db.Where("position_id = ? AND permission_id = ? AND scope_type = ?", positionID, permissionID, scopeType)
	if scopeID == nil {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", *scopeID)
	}
	return query.Delete(&models.PositionPermissionGrant{}).Error
}

// GetPositionGrants returns all scoped permissions for a position
func (r *PermissionGrantRepository) GetPositionGrants(positionID int64) ([]models.PositionPermissionGrant, error) {
	var grants []models.PositionPermissionGrant
	err := r.db.Where("position_id = ?", positionID).
		Preload("Permission").
		Find(&grants).Error
	return grants, err
}

// GetPositionGrantsByIDs returns all scoped permissions for multiple positions
func (r *PermissionGrantRepository) GetPositionGrantsByIDs(positionIDs []int64) ([]models.PositionPermissionGrant, error) {
	var grants []models.PositionPermissionGrant
	if len(positionIDs) == 0 {
		return grants, nil
	}
	err := r.db.Where("position_id IN ?", positionIDs).
		Preload("Permission").
		Find(&grants).Error
	return grants, err
}

// === Employee Permission Grants ===

// GrantToEmployee adds a scoped permission to an employee
func (r *PermissionGrantRepository) GrantToEmployee(grant *models.EmployeePermissionGrant) error {
	return r.db.Create(grant).Error
}

// RevokeFromEmployee removes a scoped permission from an employee
func (r *PermissionGrantRepository) RevokeFromEmployee(employeeID, permissionID int64, scopeType models.ScopeType, scopeID *int64) error {
	query := r.db.Where("employee_id = ? AND permission_id = ? AND scope_type = ?", employeeID, permissionID, scopeType)
	if scopeID == nil {
		query = query.Where("scope_id IS NULL")
	} else {
		query = query.Where("scope_id = ?", *scopeID)
	}
	return query.Delete(&models.EmployeePermissionGrant{}).Error
}

// GetEmployeeGrants returns all scoped permissions for an employee
func (r *PermissionGrantRepository) GetEmployeeGrants(employeeID int64) ([]models.EmployeePermissionGrant, error) {
	var grants []models.EmployeePermissionGrant
	err := r.db.Where("employee_id = ?", employeeID).
		Preload("Permission").
		Find(&grants).Error
	return grants, err
}

// GetEmployeeGrantsByIDs returns all scoped permissions for multiple employees
func (r *PermissionGrantRepository) GetEmployeeGrantsByIDs(employeeIDs []int64) ([]models.EmployeePermissionGrant, error) {
	var grants []models.EmployeePermissionGrant
	if len(employeeIDs) == 0 {
		return grants, nil
	}
	err := r.db.Where("employee_id IN ?", employeeIDs).
		Preload("Permission").
		Find(&grants).Error
	return grants, err
}

// === Scoped Permission Check ===

// CheckPositionHasScopedPermission checks if any of the positions has the permission for the given scope
func (r *PermissionGrantRepository) CheckPositionHasScopedPermission(
	positionIDs []int64,
	permissionCode string,
	scopeType models.ScopeType,
	scopeID *int64,
) (bool, error) {
	if len(positionIDs) == 0 {
		return false, nil
	}

	var count int64
	query := r.db.Model(&models.PositionPermissionGrant{}).
		Joins("JOIN permissions ON permissions.id = position_permission_grants.permission_id").
		Where("position_permission_grants.position_id IN ?", positionIDs).
		Where("permissions.code = ?", permissionCode).
		Where("(position_permission_grants.scope_type = 'organization' OR (position_permission_grants.scope_type = ? AND (position_permission_grants.scope_id IS NULL OR position_permission_grants.scope_id = ?)))", scopeType, scopeID)

	err := query.Count(&count).Error
	return count > 0, err
}

// CheckEmployeeHasScopedPermission checks if any of the employees has the permission for the given scope
func (r *PermissionGrantRepository) CheckEmployeeHasScopedPermission(
	employeeIDs []int64,
	permissionCode string,
	scopeType models.ScopeType,
	scopeID *int64,
) (bool, error) {
	if len(employeeIDs) == 0 {
		return false, nil
	}

	var count int64
	query := r.db.Model(&models.EmployeePermissionGrant{}).
		Joins("JOIN permissions ON permissions.id = employee_permission_grants.permission_id").
		Where("employee_permission_grants.employee_id IN ?", employeeIDs).
		Where("permissions.code = ?", permissionCode).
		Where("(employee_permission_grants.scope_type = 'organization' OR (employee_permission_grants.scope_type = ? AND (employee_permission_grants.scope_id IS NULL OR employee_permission_grants.scope_id = ?)))", scopeType, scopeID)

	err := query.Count(&count).Error
	return count > 0, err
}

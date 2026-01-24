package repository

import (
	"strings"

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

// CheckPositionHasScopedPermissionWithHierarchy checks if position has permission
// with full scope hierarchy support (organization → location → department)
func (r *PermissionGrantRepository) CheckPositionHasScopedPermissionWithHierarchy(
	positionIDs []int64,
	permissionCode string,
	requestedScope models.ScopeType,
	context *models.PermissionContext,
) (bool, error) {
	if len(positionIDs) == 0 {
		return false, nil
	}

	type scopeMap = map[string][]int64

	scopes := make(scopeMap)
	scopes["organizations"] = []int64{*context.OrgID}

	if context.HasLocationAccess() {
		scopes["location"] = []int64{*context.LocationID}
	}

	if context.HasDepartmentAccess() && context.DepartmentID != nil {
		parentDeptIDs, err := r.getAllParentDepartments(*context.DepartmentID)
		if err != nil {
			return false, err
		}
		scopes["department"] = parentDeptIDs
	}

	query := r.db.Model(&models.PositionPermissionGrant{}).
		Joins("JOIN permissions ON permissions.id = position_permission_grants.permission_id").
		Where("position_permission_grants.position_id IN ?", positionIDs).
		Where("permissions.code = ?", permissionCode)

	var conditions []string
	var args []interface{}

	for scopeType, ids := range scopes {
		if scopeType == "organization" {
			conditions = append(conditions, "(position_permission_grants.scope_type = ? AND (position_permission_grants.scope_id IS NULL OR position_permission_grants.scope_id = ?))")
			args = append(args, scopeType, ids[0])
		} else {
			conditions = append(conditions, "(position_permission_grants.scope_type = ? AND position_permission_grants.scope_id IN ?)")
			args = append(args, scopeType, ids)
		}
	}

	query = query.Where(strings.Join(conditions, " OR "), args...)

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// CheckEmployeeHasScopedPermissionWithHierarchy checks if employee has permission
// with full scope hierarchy support (organization → location → department)
func (r *PermissionGrantRepository) CheckEmployeeHasScopedPermissionWithHierarchy(
	employeeIDs []int64,
	permissionCode string,
	requestedScope models.ScopeType,
	context *models.PermissionContext,
) (bool, error) {
	if len(employeeIDs) == 0 {
		return false, nil
	}

	type scopeMap = map[string][]int64

	scopes := make(scopeMap)
	scopes["organizations"] = []int64{*context.OrgID}

	if context.HasLocationAccess() {
		scopes["location"] = []int64{*context.LocationID}
	}

	if context.HasDepartmentAccess() && context.DepartmentID != nil {
		parentDeptIDs, err := r.getAllParentDepartments(*context.DepartmentID)
		if err != nil {
			return false, err
		}
		scopes["department"] = parentDeptIDs
	}

	query := r.db.Model(&models.EmployeePermissionGrant{}).
		Joins("JOIN permissions ON permissions.id = employee_permission_grants.permission_id").
		Where("employee_permission_grants.position_id IN ?", employeeIDs).
		Where("permissions.code = ?", permissionCode)

	var conditions []string
	var args []interface{}

	for scopeType, ids := range scopes {
		if scopeType == "organization" {
			conditions = append(conditions, "(employee_permission_grants.scope_type = ? AND (employee_permission_grants.scope_id IS NULL OR employee_permission_grants.scope_id = ?))")
			args = append(args, scopeType, ids[0])
		} else {
			conditions = append(conditions, "(employee_permission_grants.scope_type = ? AND employee_permission_grants.scope_id IN ?)")
			args = append(args, scopeType, ids)
		}
	}

	query = query.Where(strings.Join(conditions, " OR "), args...)

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *PermissionGrantRepository) getAllParentDepartments(deptId int64) ([]int64, error) {
	var result []int64
	var parentId *int64

	currentId := deptId

	for {
		err := r.db.Model(&models.Department{}).Select("parent_id").Where("id = ?", currentId).Scan(&parentId).Error
		if err != nil {
			return nil, err
		}

		if parentId == nil {
			break
		}

		result = append(result, *parentId)
		currentId = *parentId
	}

	result = append(result, deptId)
	return result, nil
}

package models

// PermissionContext describes the hierarchical context for permission checking
// Used to validate scoped permissions against the resource hierarchy:
// Organization → Location → Department
type PermissionContext struct {
	// OrgID is always required
	OrgID *int64

	// LocationID is required when checking location-scoped permissions
	// and for validating department-scoped permissions in a location
	LocationID *int64

	// DepartmentID is required when checking department-scoped permissions
	DepartmentID *int64

	// DepartmentID is required when checking department-scoped permissions
	PositionID *int64
}

func (pc *PermissionContext) HasOrgAccess() bool {
	return pc.OrgID != nil
}

// HasLocationAccess checks if user has access at location level
func (pc *PermissionContext) HasLocationAccess() bool {
	return pc.LocationID != nil
}

// HasDepartmentAccess checks if user has access at department level
func (pc *PermissionContext) HasDepartmentAccess() bool {
	return pc.DepartmentID != nil
}

// MatchesScopeLocation checks if permission scope matches location context
func (pc *PermissionContext) MatchesScopeLocation(scopeID *int64) bool {
	if !pc.HasLocationAccess() {
		return false
	}
	if scopeID == nil {
		return true // nil means all locations
	}
	return *pc.LocationID == *scopeID
}

// MatchesScopeDepartment checks if permission scope matches department context
func (pc *PermissionContext) MatchesScopeDepartment(scopeID *int64) bool {
	if !pc.HasDepartmentAccess() {
		return false
	}
	if scopeID == nil {
		return true // nil means all departments
	}
	return *pc.DepartmentID == *scopeID
}

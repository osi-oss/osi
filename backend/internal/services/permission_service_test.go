package services

import (
	"testing"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPermissionTestDB creates an in-memory SQLite database for permission tests
func setupPermissionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.Location{},
		&models.Department{},
		&models.Position{},
		&models.Employee{},
		&models.Permission{},
		&models.PositionPermissionGrant{},
		&models.EmployeePermissionGrant{},
	)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// createPermTestUser creates a test user
func createPermTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	firstName := "Test"
	lastName := "User"
	user := &models.User{
		Email:           email,
		IsEmailVerified: true,
		FirstName:       &firstName,
		LastName:        &lastName,
		Status:          models.UserStatusActive,
	}
	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")
	return user
}

// createPermTestOrg creates a test organization with founder
func createPermTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
	org := &models.Organization{
		Name:   name,
		Status: models.OrgDraft,
	}
	err := db.Create(org).Error
	require.NoError(t, err, "failed to create test organization")

	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		IsMain:         true,
		SharePercent:   float64Ptr(100.0),
	}
	err = db.Create(founder).Error
	require.NoError(t, err, "failed to create test founder")

	err = db.Preload("Founders").First(org, org.ID).Error
	require.NoError(t, err, "failed to reload organization")

	return org
}

// createPermTestLocation creates a test location
func createPermTestLocation(t *testing.T, db *gorm.DB, orgID int64, name string) *models.Location {
	location := &models.Location{
		OrganizationID: orgID,
		Name:           name,
		Source:         "manual",
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err, "failed to create test location")
	return location
}

// createPermTestDepartment creates a test department
func createPermTestDepartment(t *testing.T, db *gorm.DB, locationID int64, name string) *models.Department {
	dept := &models.Department{
		LocationID: locationID,
		Name:       name,
	}
	err := db.Create(dept).Error
	require.NoError(t, err, "failed to create test department")
	return dept
}

// createPermTestPosition creates a test position
func createPermTestPosition(t *testing.T, db *gorm.DB, orgID int64, name string, deptID *int64) *models.Position {
	position := &models.Position{
		OrganizationID: orgID,
		DepartmentID:   deptID,
		Name:           name,
	}
	err := db.Create(position).Error
	require.NoError(t, err, "failed to create test position")
	return position
}

// createPermTestMember creates a test organization member (OrganizationMember)
func createPermTestMember(t *testing.T, db *gorm.DB, userID, orgID int64) *models.OrganizationMember {
	member := &models.OrganizationMember{
		UserID:         userID,
		OrganizationID: orgID,
		Status:         models.MemberActive,
	}
	err := db.Create(member).Error
	require.NoError(t, err, "failed to create test member")
	return member
}

// createPermTestEmployee creates a test employee (position holder within organization)
func createPermTestEmployee(t *testing.T, db *gorm.DB, memberID, positionID int64) *models.Employee {
	// First get the member to get userID and orgID
	var member models.OrganizationMember
	err := db.First(&member, memberID).Error
	require.NoError(t, err, "failed to get member")

	employee := &models.Employee{
		UserID:         member.UserID,
		OrganizationID: member.OrganizationID,
		PositionID:     positionID,
		Status:         models.MemberActive,
	}
	err = db.Create(employee).Error
	require.NoError(t, err, "failed to create test employee")
	return employee
}

// createPermTestPermission creates a test permission
func createPermTestPermission(t *testing.T, db *gorm.DB, code, description, group string) *models.Permission {
	permission := &models.Permission{
		Code:        code,
		Description: description,
		GroupName:   group,
	}
	err := db.Create(permission).Error
	require.NoError(t, err, "failed to create test permission")
	return permission
}

// setupPermissionService creates a PermissionService with all required repositories
func setupPermissionService(t *testing.T, db *gorm.DB) *PermissionService {
	permissionRepo := repository.NewPermissionRepository(db)
	permissionGrantRepo := repository.NewPermissionGrantRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	return NewPermissionService(permissionRepo, permissionGrantRepo, orgRepo, employeeRepo)
}

// TestUserHasAccessToOrganization tests access check for founders and members
func TestUserHasAccessToOrganization(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	founder := createPermTestUser(t, db, "founder@example.com")
	member := createPermTestUser(t, db, "member@example.com")
	stranger := createPermTestUser(t, db, "stranger@example.com")

	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	createPermTestMember(t, db, member.ID, org.ID)

	tests := []struct {
		name       string
		userID     int64
		orgID      int64
		wantAccess bool
	}{
		{
			name:       "Founder has access",
			userID:     founder.ID,
			orgID:      org.ID,
			wantAccess: true,
		},
		{
			name:       "Active member has access",
			userID:     member.ID,
			orgID:      org.ID,
			wantAccess: true,
		},
		{
			name:       "Stranger has no access",
			userID:     stranger.ID,
			orgID:      org.ID,
			wantAccess: false,
		},
		{
			name:       "Non-existent org returns no access",
			userID:     founder.ID,
			orgID:      99999,
			wantAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasAccess, err := svc.UserHasAccessToOrganization(tt.userID, tt.orgID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAccess, hasAccess)
		})
	}
}

// TestFounderHasAllPermissions tests that founders have all permissions
func TestFounderHasAllPermissions(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	founder := createPermTestUser(t, db, "founder@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")

	// Create some permissions
	createPermTestPermission(t, db, "locations.create", "Create locations", "Locations")
	createPermTestPermission(t, db, "departments.create", "Create departments", "Departments")

	// Founder should have all permissions without any grants
	hasPermission, err := svc.UserHasScopedPermission(founder.ID, org.ID, "locations.create", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Founder should have locations.create permission")

	hasPermission, err = svc.UserHasScopedPermission(founder.ID, org.ID, "departments.create", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Founder should have departments.create permission")

	// Even non-existent permission code should return true for founder
	hasPermission, err = svc.UserHasScopedPermission(founder.ID, org.ID, "any.permission", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Founder should have any permission")
}

// TestPositionPermissionGrant tests scoped permissions through position grants
func TestPositionPermissionGrant(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)

	// Setup location and department
	location := createPermTestLocation(t, db, org.ID, "Main Office")
	dept := createPermTestDepartment(t, db, location.ID, "Engineering")

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Developer", &dept.ID)
	createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission
	createPermTestPermission(t, db, "departments.update", "Update departments", "Departments")

	// Initially member should not have permission
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should not have permission initially")

	// Grant permission to position
	err = svc.GrantPermissionToPosition(position.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Now member should have permission
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Member should have permission after grant")

	// Revoke permission
	err = svc.RevokePermissionFromPosition(position.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Member should not have permission again
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should not have permission after revoke")
}

// TestEmployeePermissionGrant tests scoped permissions through employee grants
func TestEmployeePermissionGrant(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Developer", nil)
	employee := createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission
	createPermTestPermission(t, db, "employees.manage", "Manage employees", "Employees")

	// Initially member should not have permission
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should not have permission initially")

	// Grant permission directly to employee
	err = svc.GrantPermissionToEmployee(employee.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Now member should have permission
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Member should have permission after grant")

	// Revoke permission
	err = svc.RevokePermissionFromEmployee(employee.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Member should not have permission again
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should not have permission after revoke")
}

// TestScopedPermissionLocation tests location-scoped permissions
func TestScopedPermissionLocation(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)

	// Setup locations
	location1 := createPermTestLocation(t, db, org.ID, "Office 1")
	location2 := createPermTestLocation(t, db, org.ID, "Office 2")

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Location Manager", nil)
	createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission
	createPermTestPermission(t, db, "departments.create", "Create departments", "Departments")

	// Grant permission only for location1
	err := svc.GrantPermissionToPosition(position.ID, "departments.create", models.ScopeLocation, &location1.ID)
	require.NoError(t, err)

	// Member should have permission for location1
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.create", models.ScopeLocation, &location1.ID)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Member should have permission for location1")

	// Member should NOT have permission for location2
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.create", models.ScopeLocation, &location2.ID)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should NOT have permission for location2")
}

// TestScopedPermissionDepartment tests department-scoped permissions
func TestScopedPermissionDepartment(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)

	// Setup location and departments
	location := createPermTestLocation(t, db, org.ID, "Main Office")
	dept1 := createPermTestDepartment(t, db, location.ID, "Engineering")
	dept2 := createPermTestDepartment(t, db, location.ID, "Sales")

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Department Lead", nil)
	createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission
	createPermTestPermission(t, db, "positions.create", "Create positions", "Positions")

	// Grant permission only for dept1
	err := svc.GrantPermissionToPosition(position.ID, "positions.create", models.ScopeDepartment, &dept1.ID)
	require.NoError(t, err)

	// Member should have permission for dept1
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "positions.create", models.ScopeDepartment, &dept1.ID)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Member should have permission for dept1")

	// Member should NOT have permission for dept2
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "positions.create", models.ScopeDepartment, &dept2.ID)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Member should NOT have permission for dept2")
}

// TestOrganizationScopeCoversAll tests that organization-scoped permission covers all scopes
func TestOrganizationScopeCoversAll(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)

	// Setup location and department
	location := createPermTestLocation(t, db, org.ID, "Main Office")
	dept := createPermTestDepartment(t, db, location.ID, "Engineering")

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Admin", nil)
	createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission
	createPermTestPermission(t, db, "departments.update", "Update departments", "Departments")

	// Grant permission at organization level (no scope_id)
	err := svc.GrantPermissionToPosition(position.ID, "departments.update", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Member should have permission for any location
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.update", models.ScopeLocation, &location.ID)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Org-level permission should cover location scope")

	// Member should have permission for any department
	hasPermission, err = svc.UserHasScopedPermission(memberUser.ID, org.ID, "departments.update", models.ScopeDepartment, &dept.ID)
	require.NoError(t, err)
	assert.True(t, hasPermission, "Org-level permission should cover department scope")
}

// TestInactiveMemberNoPermission tests that inactive members have no permissions
func TestInactiveMemberNoPermission(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Setup organization
	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")

	// Create inactive member
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberUser.ID,
		Status:         models.MemberInvited, // Not active
	}
	err := db.Create(member).Error
	require.NoError(t, err)

	// Setup position and employee
	position := createPermTestPosition(t, db, org.ID, "Developer", nil)
	createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permission and grant it
	createPermTestPermission(t, db, "locations.create", "Create locations", "Locations")
	err = svc.GrantPermissionToPosition(position.ID, "locations.create", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Inactive member should not have permission
	hasPermission, err := svc.UserHasScopedPermission(memberUser.ID, org.ID, "locations.create", models.ScopeOrganization, nil)
	require.NoError(t, err)
	assert.False(t, hasPermission, "Inactive member should not have permission")
}

// TestGetPositionPermissionGrants tests retrieving all grants for a position
func TestGetPositionPermissionGrants(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	founder := createPermTestUser(t, db, "founder@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	position := createPermTestPosition(t, db, org.ID, "Admin", nil)
	location := createPermTestLocation(t, db, org.ID, "Office")

	// Create permissions
	createPermTestPermission(t, db, "locations.create", "Create locations", "Locations")
	createPermTestPermission(t, db, "departments.create", "Create departments", "Departments")

	// Grant multiple permissions
	err := svc.GrantPermissionToPosition(position.ID, "locations.create", models.ScopeOrganization, nil)
	require.NoError(t, err)
	err = svc.GrantPermissionToPosition(position.ID, "departments.create", models.ScopeLocation, &location.ID)
	require.NoError(t, err)

	// Get all grants
	grants, err := svc.GetPositionPermissionGrants(position.ID)
	require.NoError(t, err)
	assert.Len(t, grants, 2, "Should have 2 grants")
}

// TestGetEmployeePermissionGrants tests retrieving all grants for an employee
func TestGetEmployeePermissionGrants(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	founder := createPermTestUser(t, db, "founder@example.com")
	memberUser := createPermTestUser(t, db, "member@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	member := createPermTestMember(t, db, memberUser.ID, org.ID)
	position := createPermTestPosition(t, db, org.ID, "Developer", nil)
	employee := createPermTestEmployee(t, db, member.ID, position.ID)

	// Create permissions
	createPermTestPermission(t, db, "employees.view", "View employees", "Employees")
	createPermTestPermission(t, db, "employees.manage", "Manage employees", "Employees")

	// Grant multiple permissions to employee
	err := svc.GrantPermissionToEmployee(employee.ID, "employees.view", models.ScopeOrganization, nil)
	require.NoError(t, err)
	err = svc.GrantPermissionToEmployee(employee.ID, "employees.manage", models.ScopeOrganization, nil)
	require.NoError(t, err)

	// Get all grants
	grants, err := svc.GetEmployeePermissionGrants(employee.ID)
	require.NoError(t, err)
	assert.Len(t, grants, 2, "Should have 2 grants")
}

// TestGetAllPermissions tests retrieving all available permissions
func TestGetAllPermissions(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	// Create some permissions
	createPermTestPermission(t, db, "locations.create", "Create locations", "Locations")
	createPermTestPermission(t, db, "locations.update", "Update locations", "Locations")
	createPermTestPermission(t, db, "departments.create", "Create departments", "Departments")

	permissions, err := svc.GetAllPermissions()
	require.NoError(t, err)
	assert.Len(t, permissions, 3, "Should have 3 permissions")
}

// TestGetPermissionByCode tests retrieving a permission by code
func TestGetPermissionByCode(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	createPermTestPermission(t, db, "locations.create", "Create locations", "Locations")

	// Existing permission
	permission, err := svc.GetPermissionByCode("locations.create")
	require.NoError(t, err)
	assert.Equal(t, "locations.create", permission.Code)
	assert.Equal(t, "Create locations", permission.Description)

	// Non-existent permission
	_, err = svc.GetPermissionByCode("nonexistent.permission")
	require.ErrorIs(t, err, apperrors.ErrPermissionNotFound)
}

// TestGrantNonExistentPermission tests error handling for non-existent permissions
func TestGrantNonExistentPermission(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := setupPermissionService(t, db)

	founder := createPermTestUser(t, db, "founder@example.com")
	org := createPermTestOrg(t, db, founder.ID, "Test Org")
	position := createPermTestPosition(t, db, org.ID, "Developer", nil)

	// Try to grant non-existent permission
	err := svc.GrantPermissionToPosition(position.ID, "nonexistent.permission", models.ScopeOrganization, nil)
	require.ErrorIs(t, err, apperrors.ErrPermissionNotFound)
}

// ===== Helper Functions =====

// float64Ptr returns a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

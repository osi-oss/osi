package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/osi-oss/osi/internal/models"
)

/*
===========================
 MOCKS
===========================
*/

type MockPermissionRepo struct {
	mock.Mock
}

func (m *MockPermissionRepo) GetByCode(code string) (*models.Permission, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepo) GetAll() ([]models.Permission, error) {
	args := m.Called()
	return args.Get(0).([]models.Permission), args.Error(1)
}

// ------------------------

type MockPermissionGrantRepo struct {
	mock.Mock
}

func (m *MockPermissionGrantRepo) CheckPositionHasScopedPermissionWithHierarchy(
	positionIDs []int64,
	permissionCode string,
	scopeType models.ScopeType,
	context *models.PermissionContext,
) (bool, error) {
	args := m.Called(positionIDs, permissionCode, scopeType, context)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionGrantRepo) CheckEmployeeHasScopedPermissionWithHierarchy(
	employeeIDs []int64,
	permissionCode string,
	scopeType models.ScopeType,
	context *models.PermissionContext,
) (bool, error) {
	args := m.Called(employeeIDs, permissionCode, scopeType, context)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionGrantRepo) GrantToPosition(grant *models.PositionPermissionGrant) error {
	args := m.Called(grant)
	return args.Error(0)
}

func (m *MockPermissionGrantRepo) GrantToEmployee(grant *models.EmployeePermissionGrant) error {
	args := m.Called(grant)
	return args.Error(0)
}

func (m *MockPermissionGrantRepo) RevokeFromEmployee(
	employeeID int64,
	permissionID int64,
	scopeType models.ScopeType,
	scopeID *int64,
) error {
	args := m.Called(employeeID, permissionID, scopeType, scopeID)
	return args.Error(0)
}

func (m *MockPermissionGrantRepo) GetEmployeeGrants(employeeID int64) ([]models.EmployeePermissionGrant, error) {
	args := m.Called(employeeID)
	return args.Get(0).([]models.EmployeePermissionGrant), args.Error(1)
}

// ------------------------

type MockOrgRepo struct {
	mock.Mock
}

func (m *MockOrgRepo) GetFounderByUserAndOrgID(userID, orgID int64) (*models.OrganizationFounder, error) {
	args := m.Called(userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrganizationFounder), args.Error(1)
}

func (m *MockOrgRepo) GetEmployeeByUserAndOrgID(userID, orgID int64) (*models.Employee, error) {
	args := m.Called(userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockOrgRepo) GetEmployeesByUserAndOrgID(userID, orgID int64) ([]models.Employee, error) {
	args := m.Called(userID, orgID)
	return args.Get(0).([]models.Employee), args.Error(1)
}

// ------------------------

type MockEmployeeRepo struct {
	mock.Mock
}

/*
===========================
 TESTS
===========================
*/

func TestUserHasAccessToOrganization_Founder(t *testing.T) {
	orgRepo := new(MockOrgRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		int64(1),
		int64(10),
	).Return(&models.OrganizationFounder{ID: 5}, nil)

	service := NewPermissionService(
		nil,
		nil,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasAccessToOrganization(1, 10)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUserHasAccessToOrganization_ActiveEmployee(t *testing.T) {
	orgRepo := new(MockOrgRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return(nil, gorm.ErrRecordNotFound)

	orgRepo.On(
		"GetEmployeeByUserAndOrgID",
		int64(1),
		int64(10),
	).Return(&models.Employee{
		ID:     3,
		Status: models.MemberActive,
	}, nil)

	service := NewPermissionService(
		nil,
		nil,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasAccessToOrganization(1, 10)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUserHasAccessToOrganization_NoAccess(t *testing.T) {
	orgRepo := new(MockOrgRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return(nil, gorm.ErrRecordNotFound)

	orgRepo.On(
		"GetEmployeeByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return(nil, gorm.ErrRecordNotFound)

	service := NewPermissionService(
		nil,
		nil,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasAccessToOrganization(1, 10)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUserHasScopedPermissionWithHierarchy_FounderAlwaysAllowed(t *testing.T) {
	orgRepo := new(MockOrgRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		int64(1),
		int64(10),
	).Return(&models.OrganizationFounder{ID: 1}, nil)

	service := NewPermissionService(
		nil,
		nil,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasScopedPermissionWithHierarchy(
		1,
		10,
		"users.read",
		models.ScopeOrganization,
		nil,
	)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUserHasScopedPermissionWithHierarchy_ByPosition(t *testing.T) {
	orgRepo := new(MockOrgRepo)
	grantRepo := new(MockPermissionGrantRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return(nil, gorm.ErrRecordNotFound)

	orgRepo.On(
		"GetEmployeesByUserAndOrgID",
		int64(1),
		int64(10),
	).Return([]models.Employee{
		{
			ID:         1,
			PositionID: 100,
			Status:     models.MemberActive,
		},
	}, nil)

	grantRepo.On(
		"CheckPositionHasScopedPermissionWithHierarchy",
		[]int64{100},
		"users.read",
		models.ScopeOrganization,
		mock.Anything,
	).Return(true, nil)

	service := NewPermissionService(
		nil,
		grantRepo,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasScopedPermissionWithHierarchy(
		1,
		10,
		"users.read",
		models.ScopeOrganization,
		nil,
	)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUserHasScopedPermissionWithHierarchy_NoPermission(t *testing.T) {
	orgRepo := new(MockOrgRepo)
	grantRepo := new(MockPermissionGrantRepo)

	orgRepo.On(
		"GetFounderByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return(nil, gorm.ErrRecordNotFound)

	orgRepo.On(
		"GetEmployeesByUserAndOrgID",
		mock.Anything,
		mock.Anything,
	).Return([]models.Employee{}, nil)

	service := NewPermissionService(
		nil,
		grantRepo,
		orgRepo,
		nil,
	)

	ok, err := service.UserHasScopedPermissionWithHierarchy(
		1,
		10,
		"users.read",
		models.ScopeOrganization,
		nil,
	)

	assert.NoError(t, err)
	assert.False(t, ok)
}

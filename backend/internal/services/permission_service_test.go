package services

import (
	"testing"
	"time"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PermissionServiceTestSuite - тестовый набор для PermissionService
type PermissionServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	service     *PermissionService
	testOrg     *models.Organization
	testUser    *models.User
	testEmp     *models.Employee
	testPos     *models.Position
	permissions map[string]*models.Permission
}

func (s *PermissionServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	err = s.db.AutoMigrate(
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
	s.Require().NoError(err)

	// Инициализация репозиториев
	permRepo := repository.NewPermissionRepository(s.db)
	grantRepo := repository.NewPermissionGrantRepository(s.db)
	orgRepo := repository.NewOrganizationRepository(s.db)
	empRepo := repository.NewEmployeeRepository(s.db)

	s.service = NewPermissionService(permRepo, grantRepo, orgRepo, empRepo)

	// Создаём тестовые данные
	s.setupTestData()
}

func (s *PermissionServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// setupTestData создаёт базовые тестовые данные
func (s *PermissionServiceTestSuite) setupTestData() {
	// Создаём пользователя
	firstName := "Test"
	lastName := "User"
	s.testUser = &models.User{
		Email:     "test@example.com",
		FirstName: &firstName,
		LastName:  &lastName,
		Status:    models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(s.testUser).Error)

	// Создаём организацию
	s.testOrg = &models.Organization{
		Name:   "Test Org",
		Status: models.OrgDraft,
	}
	s.Require().NoError(s.db.Create(s.testOrg).Error)

	// Создаём локацию
	location := &models.Location{
		OrganizationID: s.testOrg.ID,
		Name:           "Main Office",
		Source:         "manual",
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(location).Error)

	// Создаём отдел
	department := &models.Department{
		LocationID: location.ID,
		Name:       "Engineering",
	}
	s.Require().NoError(s.db.Create(department).Error)

	// Создаём позицию
	s.testPos = &models.Position{
		OrganizationID: s.testOrg.ID,
		DepartmentID:   &department.ID,
		Name:           "Software Engineer",
		IsAdmin:        false,
	}
	s.Require().NoError(s.db.Create(s.testPos).Error)

	// Создаём сотрудника
	s.testEmp = &models.Employee{
		OrganizationID: s.testOrg.ID,
		UserID:         s.testUser.ID,
		PositionID:     s.testPos.ID,
		Status:         models.MemberActive,
	}
	s.Require().NoError(s.db.Create(s.testEmp).Error)

	// Создаём разрешения
	s.permissions = s.createTestPermissions()
}

// createTestPermissions создаёт набор тестовых разрешений
func (s *PermissionServiceTestSuite) createTestPermissions() map[string]*models.Permission {
	perms := map[string]*models.Permission{
		"users.read": {
			Code:        "users.read",
			Description: "Can view users",
		},
		"users.write": {
			Code:        "users.write",
			Description: "Can create/update users",
		},
		"org.manage": {
			Code:        "org.manage",
			Description: "Can manage organization settings",
		},
	}

	for _, perm := range perms {
		s.Require().NoError(s.db.Create(perm).Error)
	}

	return perms
}

// TestUserHasAccessToOrganization_Founder тестирует доступ основателя
func (s *PermissionServiceTestSuite) TestUserHasAccessToOrganization_Founder() {
	// Создаём основателя
	founder := &models.OrganizationFounder{
		OrganizationID: s.testOrg.ID,
		UserID:         s.testUser.ID,
		IsMain:         true,
	}
	s.Require().NoError(s.db.Create(founder).Error)

	// Проверяем доступ
	hasAccess, err := s.service.UserHasAccessToOrganization(s.testUser.ID, s.testOrg.ID)
	s.Assert().NoError(err)
	s.Assert().True(hasAccess)
}

// TestUserHasAccessToOrganization_ActiveEmployee тестирует доступ активного сотрудника
func (s *PermissionServiceTestSuite) TestUserHasAccessToOrganization_ActiveEmployee() {
	// У нас уже есть активный сотрудник из setupTestData
	hasAccess, err := s.service.UserHasAccessToOrganization(s.testUser.ID, s.testOrg.ID)
	s.Assert().NoError(err)
	s.Assert().True(hasAccess)
}

// TestUserHasAccessToOrganization_InactiveEmployee тестирует доступ неактивного сотрудника
func (s *PermissionServiceTestSuite) TestUserHasAccessToOrganization_InactiveEmployee() {
	// Деактивируем сотрудника (устанавливаем EndDate)
	endDate := time.Now()
	s.testEmp.EndDate = &endDate
	s.Require().NoError(s.db.Save(s.testEmp).Error)

	hasAccess, err := s.service.UserHasAccessToOrganization(s.testUser.ID, s.testOrg.ID)
	s.Assert().NoError(err)
	s.Assert().False(hasAccess)
}

// TestUserHasAccessToOrganization_NoAccess тестирует отсутствие доступа
func (s *PermissionServiceTestSuite) TestUserHasAccessToOrganization_NoAccess() {
	// Создаём другого пользователя
	stranger := &models.User{
		Email:  "stranger@example.com",
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(stranger).Error)

	hasAccess, err := s.service.UserHasAccessToOrganization(stranger.ID, s.testOrg.ID)
	s.Assert().NoError(err)
	s.Assert().False(hasAccess)
}

// TestFounderHasAllPermissions тестирует что основатель имеет все права
func (s *PermissionServiceTestSuite) TestFounderHasAllPermissions() {
	// Создаём основателя
	founder := &models.OrganizationFounder{
		OrganizationID: s.testOrg.ID,
		UserID:         s.testUser.ID,
		IsMain:         true,
	}
	s.Require().NoError(s.db.Create(founder).Error)

	// Проверяем различные права
	testCases := []struct {
		permission string
		scopeType  models.ScopeType
	}{
		{"users.read", models.ScopeOrganization},
		{"users.write", models.ScopeOrganization},
		{"org.manage", models.ScopeOrganization},
		{"users.read", models.ScopeLocation},
		{"users.read", models.ScopeDepartment},
	}

	for _, tc := range testCases {
		s.Run(tc.permission+"_"+string(tc.scopeType), func() {
			hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
				s.testUser.ID,
				s.testOrg.ID,
				tc.permission,
				tc.scopeType,
				nil,
			)
			s.Assert().NoError(err)
			s.Assert().True(hasPermission, "founder should have all permissions")
		})
	}
}

// TestPermissionGrantToPosition тестирует выдачу прав позиции
func (s *PermissionServiceTestSuite) TestPermissionGrantToPosition() {
	// Выдаём право позиции
	err := s.service.GrantPermissionToPosition(
		s.testPos.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)

	// Проверяем что право есть
	hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().True(hasPermission)

	// Проверяем что другого права нет
	hasPermission, err = s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.write",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().False(hasPermission)
}

// TestPermissionGrantToEmployee тестирует выдачу прав сотруднику
func (s *PermissionServiceTestSuite) TestPermissionGrantToEmployee() {
	// Выдаём право сотруднику
	err := s.service.GrantPermissionToEmployee(
		s.testEmp.ID,
		"org.manage",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)

	// Проверяем что право есть
	hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"org.manage",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().True(hasPermission)
}

// TestRevokePermissionFromEmployee тестирует отзыв прав у сотрудника
func (s *PermissionServiceTestSuite) TestRevokePermissionFromEmployee() {
	// Сначала выдаём право
	err := s.service.GrantPermissionToEmployee(
		s.testEmp.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Require().NoError(err)

	// Проверяем что право есть
	hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().True(hasPermission)

	// Отзываем право
	err = s.service.RevokePermissionFromEmployee(
		s.testEmp.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)

	// Проверяем что права больше нет
	hasPermission, err = s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().False(hasPermission)
}

// TestGetEmployeePermissionGrants тестирует получение прав сотрудника
func (s *PermissionServiceTestSuite) TestGetEmployeePermissionGrants() {
	// Выдаём несколько прав
	permissions := []string{"users.read", "users.write"}
	for _, perm := range permissions {
		err := s.service.GrantPermissionToEmployee(
			s.testEmp.ID,
			perm,
			models.ScopeOrganization,
			nil,
		)
		s.Require().NoError(err)
	}

	// Получаем список прав
	grants, err := s.service.GetEmployeePermissionGrants(s.testEmp.ID)
	s.Assert().NoError(err)
	s.Assert().Len(grants, 2)
}

// TestGetAllPermissions тестирует получение всех доступных прав
func (s *PermissionServiceTestSuite) TestGetAllPermissions() {
	perms, err := s.service.GetAllPermissions()
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(perms), 3, "should have at least the test permissions")
}

// TestGetPermissionByCode тестирует получение права по коду
func (s *PermissionServiceTestSuite) TestGetPermissionByCode() {
	// Существующее право
	perm, err := s.service.GetPermissionByCode("users.read")
	s.Assert().NoError(err)
	s.Assert().NotNil(perm)
	s.Assert().Equal("users.read", perm.Code)

	// Несуществующее право
	_, err = s.service.GetPermissionByCode("nonexistent.permission")
	s.Assert().Error(err)
}

// TestScopedPermissions тестирует иерархию scope
func (s *PermissionServiceTestSuite) TestScopedPermissions() {
	location := &models.Location{
		OrganizationID: s.testOrg.ID,
		Name:           "Office 2",
		Source:         "manual",
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(location).Error)

	department := &models.Department{
		LocationID: location.ID,
		Name:       "Sales",
	}
	s.Require().NoError(s.db.Create(department).Error)

	// Выдаём право на уровне локации
	err := s.service.GrantPermissionToPosition(
		s.testPos.ID,
		"users.read",
		models.ScopeLocation,
		&location.ID,
	)
	s.Assert().NoError(err)

	// Проверяем с правильным контекстом
	hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeLocation,
		&models.PermissionContext{
			OrgID:      &s.testOrg.ID,
			LocationID: &location.ID,
		},
	)
	s.Assert().NoError(err)
	s.Assert().True(hasPermission)

	// Проверяем с другой локацией - не должно работать
	otherLocationID := int64(9999)
	hasPermission, err = s.service.UserHasScopedPermissionWithHierarchy(
		s.testUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeLocation,
		&models.PermissionContext{
			OrgID: &s.testOrg.ID,
			LocationID: &otherLocationID,
		},
	)
	s.Assert().NoError(err)
	s.Assert().False(hasPermission)
}

// TestNoAccessWithoutEmployment тестирует отсутствие прав без трудоустройства
func (s *PermissionServiceTestSuite) TestNoAccessWithoutEmployment() {
	// Создаём нового пользователя без сотрудника
	newUser := &models.User{
		Email:  "newuser@example.com",
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(newUser).Error)

	// Проверяем что у него нет прав
	hasPermission, err := s.service.UserHasScopedPermissionWithHierarchy(
		newUser.ID,
		s.testOrg.ID,
		"users.read",
		models.ScopeOrganization,
		nil,
	)
	s.Assert().NoError(err)
	s.Assert().False(hasPermission)
}

// Запуск тестового набора
func TestPermissionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionServiceTestSuite))
}

package services

import (
	"testing"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// OrganizationServiceTestSuite - тестовый набор для OrganizationService
type OrganizationServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *OrganizationService
	orgRepo *repository.OrganizationRepository
}

// SetupTest выполняется перед каждым тестом
func (s *OrganizationServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err, "failed to open test database")

	err = s.db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.Employee{},
	)
	s.Require().NoError(err, "failed to run migrations")

	s.orgRepo = repository.NewOrganizationRepository(s.db)
	s.service = NewOrganizationService(s.orgRepo)
}

// TearDownTest выполняется после каждого теста
func (s *OrganizationServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// createTestUser создаёт тестового пользователя
func (s *OrganizationServiceTestSuite) createTestUser(email string) *models.User {
	firstName := "Test"
	lastName := "User"
	user := &models.User{
		Email:           email,
		IsEmailVerified: true,
		FirstName:       &firstName,
		LastName:        &lastName,
		Status:          models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(user).Error)
	return user
}

// TestCreateOrganization_Success тестирует успешное создание организации
func (s *OrganizationServiceTestSuite) TestCreateOrganization_Success() {
	user := s.createTestUser("test@example.com")

	testCases := []struct {
		name     string
		input    dto.CreateOrganizationRequest
		checkOrg func(*models.Organization)
	}{
		{
			name: "with all fields",
			input: dto.CreateOrganizationRequest{
				Name:         "Test Company",
				LegalName:    stringPtr("ООО Test Company"),
				INN:          stringPtr("1234567890"),
				OGRN:         stringPtr("1234567890123"),
				KPP:          stringPtr("123456789"),
				LegalAddress: stringPtr("123 Main St"),
				SharePercent: float64Ptr(100.0),
			},
			checkOrg: func(org *models.Organization) {
				s.Assert().Equal("Test Company", org.Name)
				s.Assert().Equal("ООО Test Company", *org.LegalName)
				s.Assert().Equal("1234567890", *org.INN)
				s.Assert().Equal(models.OrgDraft, org.Status)
				s.Assert().NotZero(org.ID)
				s.Assert().NotZero(org.CreatedAt)

				// Проверяем основателя
				s.Assert().Len(org.Founders, 1)
				s.Assert().Equal(user.ID, org.Founders[0].UserID)
				s.Assert().True(org.Founders[0].IsMain)
				s.Assert().Equal(100.0, *org.Founders[0].SharePercent)
			},
		},
		{
			name: "with minimal fields",
			input: dto.CreateOrganizationRequest{
				Name: "Minimal Company",
			},
			checkOrg: func(org *models.Organization) {
				s.Assert().Equal("Minimal Company", org.Name)
				s.Assert().Nil(org.LegalName)
				s.Assert().Nil(org.INN)
				s.Assert().Equal(models.OrgDraft, org.Status)
				s.Assert().Len(org.Founders, 1)
			},
		},
		{
			name: "with partial share",
			input: dto.CreateOrganizationRequest{
				Name:         "Shared Company",
				SharePercent: float64Ptr(60.0),
			},
			checkOrg: func(org *models.Organization) {
				s.Assert().Equal(60.0, *org.Founders[0].SharePercent)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			org, err := s.service.CreateOrganization(user.ID, tc.input)
			s.Require().NoError(err)
			s.Assert().NotNil(org)
			tc.checkOrg(org)
		})
	}
}

// TestCreateOrganization_Validation тестирует валидацию при создании
func (s *OrganizationServiceTestSuite) TestCreateOrganization_Validation() {
	user := s.createTestUser("test@example.com")

	testCases := []struct {
		name      string
		input     dto.CreateOrganizationRequest
		expectErr bool
	}{
		{
			name:      "empty name",
			input:     dto.CreateOrganizationRequest{Name: ""},
			expectErr: true,
		},
		{
			name:      "whitespace name",
			input:     dto.CreateOrganizationRequest{Name: "   "},
			expectErr: false, // будет создана с пробелами, бизнес-логика решает нужна ли trim
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			org, err := s.service.CreateOrganization(user.ID, tc.input)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Nil(org)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

// TestGetOrganization_Access тестирует права доступа при получении организации
func (s *OrganizationServiceTestSuite) TestGetOrganization_Access() {
	user1 := s.createTestUser("user1@example.com")
	user2 := s.createTestUser("user2@example.com")

	// user1 создаёт организацию
	org, err := s.service.CreateOrganization(user1.ID, dto.CreateOrganizationRequest{
		Name: "Company A",
	})
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		userID    int64
		orgID     int64
		expectErr bool
		errorType error
	}{
		{
			name:      "founder can access",
			userID:    user1.ID,
			orgID:     org.ID,
			expectErr: false,
		},
		{
			name:      "non-founder cannot access",
			userID:    user2.ID,
			orgID:     org.ID,
			expectErr: true,
			errorType: apperrors.ErrForbidden,
		},
		{
			name:      "organization not found",
			userID:    user1.ID,
			orgID:     999,
			expectErr: true,
			errorType: apperrors.ErrOrganizationNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.service.GetOrganization(tc.orgID, tc.userID)

			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Nil(result)
				if tc.errorType != nil {
					s.Assert().ErrorIs(err, tc.errorType)
				}
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(result)
				s.Assert().Equal(org.ID, result.ID)
			}
		})
	}
}

// TestGetOrganization_EmployeeAccess тестирует доступ сотрудников
func (s *OrganizationServiceTestSuite) TestGetOrganization_EmployeeAccess() {
	founder := s.createTestUser("founder@example.com")
	employee := s.createTestUser("employee@example.com")

	// Создаём организацию
	org, err := s.service.CreateOrganization(founder.ID, dto.CreateOrganizationRequest{
		Name: "Company with Employee",
	})
	s.Require().NoError(err)

	// Добавляем сотрудника
	emp := &models.Employee{
		OrganizationID: org.ID,
		UserID:         employee.ID,
		PositionID:     1, // dummy position
		Status:         models.MemberActive,
	}
	s.Require().NoError(s.db.Create(emp).Error)

	// Активный сотрудник может получить доступ
	result, err := s.service.GetOrganization(org.ID, employee.ID)
	s.Assert().NoError(err)
	s.Assert().NotNil(result)

	// Неактивный сотрудник не может получить доступ
	emp.Status = models.MemberBlocked
	s.Require().NoError(s.db.Save(emp).Error)

	_, err = s.service.GetOrganization(org.ID, employee.ID)
	s.Assert().Error(err)
}

// TestGetUserOrganizations тестирует получение организаций пользователя
func (s *OrganizationServiceTestSuite) TestGetUserOrganizations() {
	user := s.createTestUser("test@example.com")

	// Создаём несколько организаций
	org1, err := s.service.CreateOrganization(user.ID, dto.CreateOrganizationRequest{
		Name: "Company 1",
	})
	s.Require().NoError(err)

	org2, err := s.service.CreateOrganization(user.ID, dto.CreateOrganizationRequest{
		Name: "Company 2",
	})
	s.Require().NoError(err)

	org3, err := s.service.CreateOrganization(user.ID, dto.CreateOrganizationRequest{
		Name: "Company 3",
	})
	s.Require().NoError(err)

	// Получаем организации
	orgs, err := s.service.GetUserOrganizations(user.ID)
	s.Assert().NoError(err)
	s.Assert().Len(orgs, 3)

	// Проверяем что все организации в списке
	orgIDs := make(map[int64]bool)
	for _, org := range orgs {
		orgIDs[org.ID] = true
	}

	s.Assert().True(orgIDs[org1.ID])
	s.Assert().True(orgIDs[org2.ID])
	s.Assert().True(orgIDs[org3.ID])
}

// TestUpdateOrganization тестирует обновление организации
func (s *OrganizationServiceTestSuite) TestUpdateOrganization() {
	user1 := s.createTestUser("user1@example.com")
	user2 := s.createTestUser("user2@example.com")

	org, err := s.service.CreateOrganization(user1.ID, dto.CreateOrganizationRequest{
		Name: "Original Name",
	})
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		userID    int64
		orgID     int64
		input     dto.UpdateOrganizationRequest
		expectErr bool
		errorType error
		checkOrg  func(*models.Organization)
	}{
		{
			name:   "founder can update",
			userID: user1.ID,
			orgID:  org.ID,
			input: dto.UpdateOrganizationRequest{
				Name:      "Updated Name",
				LegalName: stringPtr("Updated Legal Name"),
				INN:       stringPtr("9876543210"),
			},
			expectErr: false,
			checkOrg: func(org *models.Organization) {
				s.Assert().Equal("Updated Name", org.Name)
				s.Assert().Equal("Updated Legal Name", *org.LegalName)
				s.Assert().Equal("9876543210", *org.INN)
			},
		},
		{
			name:      "non-founder cannot update",
			userID:    user2.ID,
			orgID:     org.ID,
			input:     dto.UpdateOrganizationRequest{Name: "Hacked"},
			expectErr: true,
			errorType: apperrors.ErrForbidden,
		},
		{
			name:      "organization not found",
			userID:    user1.ID,
			orgID:     999,
			input:     dto.UpdateOrganizationRequest{Name: "Whatever"},
			expectErr: true,
			errorType: apperrors.ErrOrganizationNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.service.UpdateOrganization(tc.orgID, tc.userID, tc.input)

			if tc.expectErr {
				s.Assert().Error(err)
				if tc.errorType != nil {
					s.Assert().ErrorIs(err, tc.errorType)
				}
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(result)
				if tc.checkOrg != nil {
					tc.checkOrg(result)
				}
			}
		})
	}
}

// TestDeleteOrganization тестирует удаление организации
func (s *OrganizationServiceTestSuite) TestDeleteOrganization() {
	user1 := s.createTestUser("user1@example.com")
	user2 := s.createTestUser("user2@example.com")

	testCases := []struct {
		name      string
		setupOrg  func() int64
		userID    int64
		expectErr bool
		errorType error
	}{
		{
			name: "main founder can delete",
			setupOrg: func() int64 {
				org, err := s.service.CreateOrganization(user1.ID, dto.CreateOrganizationRequest{
					Name: "To Delete 1",
				})
				s.Require().NoError(err)
				return org.ID
			},
			userID:    user1.ID,
			expectErr: false,
		},
		{
			name: "non-founder cannot delete",
			setupOrg: func() int64 {
				org, err := s.service.CreateOrganization(user1.ID, dto.CreateOrganizationRequest{
					Name: "To Delete 2",
				})
				s.Require().NoError(err)
				return org.ID
			},
			userID:    user2.ID,
			expectErr: true,
			errorType: apperrors.ErrForbidden,
		},
		{
			name: "organization not found",
			setupOrg: func() int64 {
				return 999
			},
			userID:    user1.ID,
			expectErr: true,
			errorType: apperrors.ErrOrganizationNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			orgID := tc.setupOrg()
			err := s.service.DeleteOrganization(orgID, tc.userID)

			if tc.expectErr {
				s.Assert().Error(err)
				if tc.errorType != nil {
					s.Assert().ErrorIs(err, tc.errorType)
				}
			} else {
				s.Assert().NoError(err)

				// Проверяем что организация удалена
				var deletedOrg models.Organization
				err := s.db.First(&deletedOrg, orgID).Error
				s.Assert().Error(err)
				s.Assert().ErrorIs(err, gorm.ErrRecordNotFound)
			}
		})
	}
}

// TestUserHasAccessToOrganization тестирует проверку доступа
func (s *OrganizationServiceTestSuite) TestUserHasAccessToOrganization() {
	founder := s.createTestUser("founder@example.com")
	employee := s.createTestUser("employee@example.com")
	stranger := s.createTestUser("stranger@example.com")

	org, err := s.service.CreateOrganization(founder.ID, dto.CreateOrganizationRequest{
		Name: "Access Test Org",
	})
	s.Require().NoError(err)

	// Добавляем активного сотрудника
	emp := &models.Employee{
		OrganizationID: org.ID,
		UserID:         employee.ID,
		PositionID:     1,
		Status:         models.MemberActive,
	}
	s.Require().NoError(s.db.Create(emp).Error)

	testCases := []struct {
		name           string
		userID         int64
		expectedAccess bool
	}{
		{"founder has access", founder.ID, true},
		{"active employee has access", employee.ID, true},
		{"stranger has no access", stranger.ID, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			hasAccess, err := s.service.UserHasAccessToOrganization(tc.userID, org.ID)
			s.Assert().NoError(err)
			s.Assert().Equal(tc.expectedAccess, hasAccess)
		})
	}
}

// TestOrganizationStatus тестирует статусы организации
func (s *OrganizationServiceTestSuite) TestOrganizationStatus() {
	user := s.createTestUser("test@example.com")

	// При создании статус должен быть draft
	org, err := s.service.CreateOrganization(user.ID, dto.CreateOrganizationRequest{
		Name: "Status Test Org",
	})
	s.Assert().NoError(err)
	s.Assert().Equal(models.OrgDraft, org.Status)

	// Проверяем что статус сохраняется
	retrieved, err := s.service.GetOrganization(org.ID, user.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(models.OrgDraft, retrieved.Status)
}

// TestMultipleFounders тестирует сценарий с несколькими основателями
func (s *OrganizationServiceTestSuite) TestMultipleFounders() {
	user1 := s.createTestUser("user1@example.com")
	user2 := s.createTestUser("user2@example.com")

	// user1 создаёт организацию с 60% долей
	org, err := s.service.CreateOrganization(user1.ID, dto.CreateOrganizationRequest{
		Name:         "Multi-founder Org",
		SharePercent: float64Ptr(60.0),
	})
	s.Require().NoError(err)
	s.Assert().Len(org.Founders, 1)

	// user1 может получить доступ
	_, err = s.service.GetOrganization(org.ID, user1.ID)
	s.Assert().NoError(err)

	// user2 не может получить доступ
	_, err = s.service.GetOrganization(org.ID, user2.ID)
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, apperrors.ErrForbidden)

	// user1 может удалить
	err = s.service.DeleteOrganization(org.ID, user1.ID)
	s.Assert().NoError(err)
}

// Запуск тестового набора
func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}

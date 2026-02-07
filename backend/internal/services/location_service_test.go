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

// LocationServiceTestSuite - тестовый набор для LocationService
type LocationServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	service      *LocationService
	locationRepo *repository.LocationRepository
	orgRepo      *repository.OrganizationRepository
	testOrg      *models.Organization
	testUser     *models.User
}

// SetupTest выполняется перед каждым тестом
func (s *LocationServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err)

	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	err = s.db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.Location{},
		&models.Department{},
		&models.Employee{},
	)
	s.Require().NoError(err)

	s.locationRepo = repository.NewLocationRepository(s.db)
	s.orgRepo = repository.NewOrganizationRepository(s.db)
	s.service = NewLocationService(s.locationRepo, s.orgRepo)

	// Создаём тестовые данные
	s.testUser = s.createTestUser("founder@example.com")
	s.testOrg = s.createTestOrg(s.testUser.ID, "Test Org")
}

// TearDownTest выполняется после каждого теста
func (s *LocationServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// createTestUser создаёт тестового пользователя
func (s *LocationServiceTestSuite) createTestUser(email string) *models.User {
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

// createTestOrg создаёт тестовую организацию
func (s *LocationServiceTestSuite) createTestOrg(userID int64, name string) *models.Organization {
	org := &models.Organization{
		Name:   name,
		Status: models.OrgDraft,
	}
	s.Require().NoError(s.db.Create(org).Error)

	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		IsMain:         true,
		SharePercent:   float64Ptr(100.0),
	}
	s.Require().NoError(s.db.Create(founder).Error)

	s.Require().NoError(s.db.Preload("Founders").First(org, org.ID).Error)
	return org
}

// TestCreateLocation_Success тестирует успешное создание локации
func (s *LocationServiceTestSuite) TestCreateLocation_Success() {
	testCases := []struct {
		name     string
		input    dto.CreateLocationRequest
		checkLoc func(*models.Location)
	}{
		{
			name: "with all fields",
			input: dto.CreateLocationRequest{
				Name:       "Main Office",
				Address:    stringPtr("123 Main St, Moscow"),
				Source:     "manual",
				IsVerified: false,
			},
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal("Main Office", loc.Name)
				s.Assert().NotNil(loc.Address)
				if loc.Address != nil {
					s.Assert().Equal("123 Main St, Moscow", *loc.Address)
				}
				s.Assert().Equal("manual", loc.Source)
				s.Assert().False(loc.IsVerified)
				s.Assert().True(loc.IsActive)
				s.Assert().Equal(s.testOrg.ID, loc.OrganizationID)
				s.Assert().NotZero(loc.ID)
				s.Assert().False(loc.CreatedAt.IsZero())
			},
		},
		{
			name: "without address",
			input: dto.CreateLocationRequest{
				Name:       "Branch Office",
				Address:    nil,
				Source:     "registry",
				IsVerified: true,
			},
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal("Branch Office", loc.Name)
				s.Assert().Nil(loc.Address)
				s.Assert().Equal("registry", loc.Source)
				s.Assert().True(loc.IsVerified)
				s.Assert().True(loc.IsActive)
			},
		},
		{
			name: "minimal fields",
			input: dto.CreateLocationRequest{
				Name:   "Simple Location",
				Source: "manual",
			},
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal("Simple Location", loc.Name)
				s.Assert().Nil(loc.Address)
				s.Assert().False(loc.IsVerified)
				s.Assert().True(loc.IsActive)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			location, err := s.service.CreateLocation(s.testOrg.ID, tc.input)
			s.Require().NoError(err)
			s.Require().NotNil(location)
			if tc.checkLoc != nil {
				tc.checkLoc(location)
			}
		})
	}
}

// TestCreateLocation_OrganizationNotFound тестирует создание для несуществующей организации
func (s *LocationServiceTestSuite) TestCreateLocation_OrganizationNotFound() {
	input := dto.CreateLocationRequest{
		Name:   "Orphan Location",
		Source: "manual",
	}

	location, err := s.service.CreateLocation(99999, input)
	s.Assert().Error(err)
	s.Assert().Nil(location)
	s.Assert().ErrorIs(err, apperrors.ErrOrganizationNotFound)
}

// TestGetLocation тестирует получение локации по ID
func (s *LocationServiceTestSuite) TestGetLocation() {
	// Создаём локацию
	location := &models.Location{
		OrganizationID: s.testOrg.ID,
		Name:           "Test Location",
		Address:        stringPtr("123 Test St"),
		Source:         "manual",
		IsVerified:     false,
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(location).Error)

	testCases := []struct {
		name       string
		locationID int64
		expectErr  bool
		checkLoc   func(*models.Location)
	}{
		{
			name:       "existing location",
			locationID: location.ID,
			expectErr:  false,
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal(location.ID, loc.ID)
				s.Assert().Equal("Test Location", loc.Name)
				s.Assert().Equal(s.testOrg.ID, loc.OrganizationID)
			},
		},
		{
			name:       "non-existent location",
			locationID: 99999,
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			loc, err := s.service.GetLocation(tc.locationID)

			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Nil(loc)
				s.Assert().ErrorIs(err, apperrors.ErrLocationNotFound)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(loc)
				if tc.checkLoc != nil {
					tc.checkLoc(loc)
				}
			}
		})
	}
}

// TestGetOrganizationLocations тестирует получение всех локаций организации
func (s *LocationServiceTestSuite) TestGetOrganizationLocations() {
	// Создаём несколько локаций
	locations := []models.Location{
		{
			OrganizationID: s.testOrg.ID,
			Name:           "Location 1",
			Source:         "manual",
			IsActive:       true,
		},
		{
			OrganizationID: s.testOrg.ID,
			Name:           "Location 2",
			Source:         "registry",
			IsActive:       true,
		},
		{
			OrganizationID: s.testOrg.ID,
			Name:           "Location 3",
			Source:         "manual",
			IsActive:       false,
		},
	}
	for i := range locations {
		s.Require().NoError(s.db.Create(&locations[i]).Error)
	}

	testCases := []struct {
		name      string
		orgID     int64
		checkLocs func([]models.Location)
	}{
		{
			name:  "all locations for organization",
			orgID: s.testOrg.ID,
			checkLocs: func(locs []models.Location) {
				s.Assert().Len(locs, 3)
				names := make(map[string]bool)
				for _, loc := range locs {
					names[loc.Name] = true
				}
				s.Assert().True(names["Location 1"])
				s.Assert().True(names["Location 2"])
				s.Assert().True(names["Location 3"])
			},
		},
		{
			name:  "empty list for organization without locations",
			orgID: 99999,
			checkLocs: func(locs []models.Location) {
				s.Assert().Len(locs, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			locs, err := s.service.GetOrganizationLocations(tc.orgID)
			s.Assert().NoError(err)
			s.Assert().NotNil(locs)
			tc.checkLocs(locs)
		})
	}
}

// TestUpdateLocation тестирует обновление локации
func (s *LocationServiceTestSuite) TestUpdateLocation() {
	// Создаём локацию для обновления
	location := &models.Location{
		OrganizationID: s.testOrg.ID,
		Name:           "Old Name",
		Address:        stringPtr("Old Address"),
		Source:         "manual",
		IsVerified:     false,
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(location).Error)

	testCases := []struct {
		name       string
		locationID int64
		input      dto.UpdateLocationRequest
		expectErr  bool
		checkLoc   func(*models.Location)
	}{
		{
			name:       "update all fields",
			locationID: location.ID,
			input: dto.UpdateLocationRequest{
				Name:       "New Name",
				Address:    stringPtr("New Address"),
				Source:     "registry",
				IsVerified: boolPtr(true),
				IsActive:   boolPtr(false),
			},
			expectErr: false,
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal("New Name", loc.Name)
				s.Assert().Equal("New Address", *loc.Address)
				s.Assert().Equal("registry", loc.Source)
				s.Assert().True(loc.IsVerified)
				s.Assert().False(loc.IsActive)
			},
		},
		{
			name:       "partial update (only name)",
			locationID: location.ID,
			input: dto.UpdateLocationRequest{
				Name: "Partially Updated",
			},
			expectErr: false,
			checkLoc: func(loc *models.Location) {
				s.Assert().Equal("Partially Updated", loc.Name)
			},
		},
		{
			name:       "update verification status",
			locationID: location.ID,
			input: dto.UpdateLocationRequest{
				IsVerified: boolPtr(true),
			},
			expectErr: false,
			checkLoc: func(loc *models.Location) {
				s.Assert().True(loc.IsVerified)
			},
		},
		{
			name:       "location not found",
			locationID: 99999,
			input: dto.UpdateLocationRequest{
				Name: "Nonexistent",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			loc, err := s.service.UpdateLocation(tc.locationID, tc.input)

			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Nil(loc)
				s.Assert().ErrorIs(err, apperrors.ErrLocationNotFound)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(loc)
				if tc.checkLoc != nil {
					tc.checkLoc(loc)
				}
			}
		})
	}
}

// TestDeleteLocation тестирует удаление локации
func (s *LocationServiceTestSuite) TestDeleteLocation() {
	testCases := []struct {
		name      string
		setupLoc  func() int64
		expectErr bool
	}{
		{
			name: "delete existing location",
			setupLoc: func() int64 {
				loc := &models.Location{
					OrganizationID: s.testOrg.ID,
					Name:           "To Delete",
					Source:         "manual",
					IsActive:       true,
				}
				s.Require().NoError(s.db.Create(loc).Error)
				return loc.ID
			},
			expectErr: false,
		},
		{
			name: "delete non-existent location",
			setupLoc: func() int64 {
				return 99999
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			locID := tc.setupLoc()
			err := s.service.DeleteLocation(locID)

			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().ErrorIs(err, apperrors.ErrLocationNotFound)
			} else {
				s.Assert().NoError(err)

				// Проверяем что локация действительно удалена
				var deletedLoc models.Location
				err := s.db.First(&deletedLoc, locID).Error
				s.Assert().Error(err)
				s.Assert().ErrorIs(err, gorm.ErrRecordNotFound)
			}
		})
	}
}

// TestLocationLifecycle тестирует полный жизненный цикл локации
func (s *LocationServiceTestSuite) TestLocationLifecycle() {
	// 1. Создание
	created, err := s.service.CreateLocation(s.testOrg.ID, dto.CreateLocationRequest{
		Name:       "Lifecycle Location",
		Address:    stringPtr("123 Start St"),
		Source:     "manual",
		IsVerified: false,
	})
	s.Require().NoError(err)
	s.Assert().False(created.IsVerified)
	s.Assert().True(created.IsActive)

	// 2. Получение
	retrieved, err := s.service.GetLocation(created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, retrieved.ID)

	// 3. Обновление
	updated, err := s.service.UpdateLocation(created.ID, dto.UpdateLocationRequest{
		Name:       "Updated Lifecycle Location",
		Address:    stringPtr("456 New St"),
		IsVerified: boolPtr(true),
	})
	s.Require().NoError(err)
	s.Assert().Equal("Updated Lifecycle Location", updated.Name)
	s.Assert().True(updated.IsVerified)

	// 4. Деактивация
	deactivated, err := s.service.UpdateLocation(created.ID, dto.UpdateLocationRequest{
		IsActive: boolPtr(false),
	})
	s.Require().NoError(err)
	s.Assert().False(deactivated.IsActive)

	// 5. Удаление
	err = s.service.DeleteLocation(created.ID)
	s.Require().NoError(err)

	// 6. Проверка что удалена
	_, err = s.service.GetLocation(created.ID)
	s.Assert().Error(err)
}

// Запуск тестового набора
func TestLocationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LocationServiceTestSuite))
}

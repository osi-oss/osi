package services

import (
	"testing"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupLocationTestDB создает тестовую базу данных в памяти для локаций
func setupLocationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	// Миграции
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.OrganizationMember{},
		&models.Location{},
		&models.Department{},
	)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// createLocationTestUser создает тестового пользователя
func createLocationTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:           stringPtr(email),
		PasswordHash:    "hashed_password",
		IsEmailVerified: true,
		FirstName:       "Test",
		LastName:        "User",
	}
	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")
	return user
}

// createLocationTestOrg создает тестовую организацию
func createLocationTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
	org := &models.Organization{
		Name:   name,
		Status: models.OrgDraft,
	}
	err := db.Create(org).Error
	require.NoError(t, err, "failed to create test organization")

	// Создаем основателя
	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		IsMain:         true,
		SharePercent:   float64Ptr(100.0),
	}
	err = db.Create(founder).Error
	require.NoError(t, err, "failed to create test founder")

	// Загружаем организацию с основателями
	err = db.Preload("Founders").First(org, org.ID).Error
	require.NoError(t, err, "failed to reload organization")

	return org
}

// TestCreateLocation тестирует создание локации
func TestCreateLocation(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	// departmentRepo := repository.NewDepartmentRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	// departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	founder := createLocationTestUser(t, db, "founder@example.com")
	nonFounder := createLocationTestUser(t, db, "nonfounder@example.com")
	org := createLocationTestOrg(t, db, founder.ID, "Test Org")

	tests := []struct {
		name        string
		orgID       int64
		userID      int64
		input       dto.CreateLocationRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkLoc    func(*testing.T, *models.Location)
	}{
		{
			name:   "Success - Create location by founder",
			orgID:  org.ID,
			userID: founder.ID,
			input: dto.CreateLocationRequest{
				Name:       "Main Office",
				Address:    stringPtr("123 Main St, Moscow"),
				Source:     "manual",
				IsVerified: false,
			},
			expectError: false,
			checkLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, "Main Office", loc.Name)
				assert.Equal(t, "123 Main St, Moscow", *loc.Address)
				assert.Equal(t, "manual", loc.Source)
				assert.False(t, loc.IsVerified)
				assert.True(t, loc.IsActive)
				assert.Equal(t, org.ID, loc.OrganizationID)
				assert.NotZero(t, loc.ID)
				assert.NotZero(t, loc.CreatedAt)
			},
		},
		{
			name:   "Success - Create location without address",
			orgID:  org.ID,
			userID: founder.ID,
			input: dto.CreateLocationRequest{
				Name:       "Branch Office",
				Address:    nil,
				Source:     "registry",
				IsVerified: true,
			},
			expectError: false,
			checkLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, "Branch Office", loc.Name)
				assert.Nil(t, loc.Address)
				assert.Equal(t, "registry", loc.Source)
				assert.True(t, loc.IsVerified)
				assert.True(t, loc.IsActive)
			},
		},
		{
			name:   "Error - Non-founder cannot create location",
			orgID:  org.ID,
			userID: nonFounder.ID,
			input: dto.CreateLocationRequest{
				Name:       "Unauthorized Location",
				Address:    stringPtr("456 Side St"),
				Source:     "manual",
				IsVerified: false,
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
		{
			name:   "Error - Organization not found",
			orgID:  99999,
			userID: founder.ID,
			input: dto.CreateLocationRequest{
				Name:       "Nonexistent Org Location",
				Address:    stringPtr("789 Wrong St"),
				Source:     "manual",
				IsVerified: false,
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				// When organization doesn't exist, user has no access, so apperrors.ErrAccessDenied is returned
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := locationService.CreateLocation(tt.orgID, tt.userID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, location)
				if tt.checkLoc != nil {
					tt.checkLoc(t, location)
				}
			}
		})
	}
}

// TestGetLocation тестирует получение локации по ID
func TestGetLocation(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	// departmentRepo := repository.NewDepartmentRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	// departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	founder := createLocationTestUser(t, db, "founder@example.com")
	nonFounder := createLocationTestUser(t, db, "nonfounder@example.com")
	org := createLocationTestOrg(t, db, founder.ID, "Test Org")

	// Создаем локацию
	location := &models.Location{
		OrganizationID: org.ID,
		Name:           "Test Location",
		Address:        stringPtr("123 Test St"),
		Source:         "manual",
		IsVerified:     false,
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		locationID  int64
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
		checkLoc    func(*testing.T, *models.Location)
	}{
		{
			name:        "Success - Get location by founder",
			locationID:  location.ID,
			userID:      founder.ID,
			expectError: false,
			checkLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, location.ID, loc.ID)
				assert.Equal(t, "Test Location", loc.Name)
				assert.Equal(t, org.ID, loc.OrganizationID)
			},
		},
		{
			name:        "Error - Non-founder cannot get location",
			locationID:  location.ID,
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
		{
			name:        "Error - Location not found",
			locationID:  99999,
			userID:      founder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrLocationNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := locationService.GetLocation(tt.locationID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, loc)
				if tt.checkLoc != nil {
					tt.checkLoc(t, loc)
				}
			}
		})
	}
}

// TestGetOrganizationLocations тестирует получение всех локаций организации
func TestGetOrganizationLocations(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	// departmentRepo := repository.NewDepartmentRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	// departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	founder := createLocationTestUser(t, db, "founder@example.com")
	nonFounder := createLocationTestUser(t, db, "nonfounder@example.com")
	org := createLocationTestOrg(t, db, founder.ID, "Test Org")

	// Создаем несколько локаций
	locations := []models.Location{
		{
			OrganizationID: org.ID,
			Name:           "Location 1",
			Source:         "manual",
			IsActive:       true,
		},
		{
			OrganizationID: org.ID,
			Name:           "Location 2",
			Source:         "registry",
			IsActive:       true,
		},
		{
			OrganizationID: org.ID,
			Name:           "Location 3",
			Source:         "manual",
			IsActive:       false,
		},
	}
	for _, loc := range locations {
		err := db.Create(&loc).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		orgID       int64
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
		checkLocs   func(*testing.T, []models.Location)
	}{
		{
			name:        "Success - Get all locations by founder",
			orgID:       org.ID,
			userID:      founder.ID,
			expectError: false,
			checkLocs: func(t *testing.T, locs []models.Location) {
				assert.Len(t, locs, 3)
				names := []string{locs[0].Name, locs[1].Name, locs[2].Name}
				assert.Contains(t, names, "Location 1")
				assert.Contains(t, names, "Location 2")
				assert.Contains(t, names, "Location 3")
			},
		},
		{
			name:        "Error - Non-founder cannot get locations",
			orgID:       org.ID,
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
		{
			name:        "Error - Organization not found",
			orgID:       99999,
			userID:      founder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				// When organization doesn't exist, UserHasAccessToOrganization returns false -> apperrors.ErrAccessDenied
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locs, err := locationService.GetOrganizationLocations(tt.orgID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, locs)
				if tt.checkLocs != nil {
					tt.checkLocs(t, locs)
				}
			}
		})
	}
}

// TestUpdateLocation тестирует обновление локации
func TestUpdateLocation(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	// departmentRepo := repository.NewDepartmentRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	// departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	founder := createLocationTestUser(t, db, "founder@example.com")
	nonFounder := createLocationTestUser(t, db, "nonfounder@example.com")
	org := createLocationTestOrg(t, db, founder.ID, "Test Org")

	// Создаем локацию
	location := &models.Location{
		OrganizationID: org.ID,
		Name:           "Old Name",
		Address:        stringPtr("Old Address"),
		Source:         "manual",
		IsVerified:     false,
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		locationID  int64
		userID      int64
		input       dto.UpdateLocationRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkLoc    func(*testing.T, *models.Location)
	}{
		{
			name:       "Success - Update all fields",
			locationID: location.ID,
			userID:     founder.ID,
			input: dto.UpdateLocationRequest{
				Name:       "New Name",
				Address:    stringPtr("New Address"),
				Source:     "registry",
				IsVerified: boolPtr(true),
				IsActive:   boolPtr(false),
			},
			expectError: false,
			checkLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, "New Name", loc.Name)
				assert.Equal(t, "New Address", *loc.Address)
				assert.Equal(t, "registry", loc.Source)
				assert.True(t, loc.IsVerified)
				assert.False(t, loc.IsActive)
			},
		},
		{
			name:       "Success - Partial update (only name)",
			locationID: location.ID,
			userID:     founder.ID,
			input: dto.UpdateLocationRequest{
				Name: "Partially Updated",
			},
			expectError: false,
			checkLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, "Partially Updated", loc.Name)
				// Другие поля должны остаться без изменений
			},
		},
		{
			name:       "Error - Non-founder cannot update",
			locationID: location.ID,
			userID:     nonFounder.ID,
			input: dto.UpdateLocationRequest{
				Name: "Unauthorized Update",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
		{
			name:       "Error - Location not found",
			locationID: 99999,
			userID:     founder.ID,
			input: dto.UpdateLocationRequest{
				Name: "Nonexistent",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrLocationNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := locationService.UpdateLocation(tt.locationID, tt.userID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, loc)
				if tt.checkLoc != nil {
					tt.checkLoc(t, loc)
				}
			}
		})
	}
}

// TestDeleteLocation тестирует удаление локации
func TestDeleteLocation(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	// departmentRepo := repository.NewDepartmentRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	// departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	founder := createLocationTestUser(t, db, "founder@example.com")
	nonFounder := createLocationTestUser(t, db, "nonfounder@example.com")
	org := createLocationTestOrg(t, db, founder.ID, "Test Org")

	tests := []struct {
		name        string
		setupLoc    func() *models.Location
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "Success - Delete location by founder",
			setupLoc: func() *models.Location {
				loc := &models.Location{
					OrganizationID: org.ID,
					Name:           "To Delete",
					Source:         "manual",
					IsActive:       true,
				}
				err := db.Create(loc).Error
				require.NoError(t, err)
				return loc
			},
			userID:      founder.ID,
			expectError: false,
		},
		{
			name: "Error - Non-founder cannot delete",
			setupLoc: func() *models.Location {
				loc := &models.Location{
					OrganizationID: org.ID,
					Name:           "Protected Location",
					Source:         "manual",
					IsActive:       true,
				}
				err := db.Create(loc).Error
				require.NoError(t, err)
				return loc
			},
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAccessDenied)
			},
		},
		{
			name: "Error - Location not found",
			setupLoc: func() *models.Location {
				// Возвращаем локацию с несуществующим ID, но не создаём её в БД
				return &models.Location{}
			},
			userID:      founder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrLocationNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := tt.setupLoc()
			// Для несуществующей локации используем ID 99999
			locID := loc.ID
			if locID == 0 {
				locID = 99999
			}
			err := locationService.DeleteLocation(locID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)

				// Проверяем, что локация действительно удалена
				var deletedLoc models.Location
				err := db.First(&deletedLoc, loc.ID).Error
				assert.Error(t, err)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

// boolPtr возвращает указатель на bool
func boolPtr(b bool) *bool {
	return &b
}

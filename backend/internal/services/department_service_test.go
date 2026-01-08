package services

import (
	"testing"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDepartmentTestDB создает тестовую базу данных для отделов
func setupDepartmentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.OrganizationMember{},
		&models.Location{},
		&models.Department{},
		&models.Position{},
	)
	require.NoError(t, err)

	return db
}

// createDepartmentTestUser создает тестового пользователя
func createDepartmentTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:           stringPtr(email),
		PasswordHash:    "hashed_password",
		IsEmailVerified: true,
		FirstName:       "Test",
		LastName:        "User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// createDepartmentTestOrg создает тестовую организацию
func createDepartmentTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
	org := &models.Organization{
		Name:   name,
		Status: models.OrgDraft,
	}
	err := db.Create(org).Error
	require.NoError(t, err)

	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		IsMain:         true,
		SharePercent:   float64Ptr(100.0),
	}
	err = db.Create(founder).Error
	require.NoError(t, err)

	err = db.Preload("Founders").First(org, org.ID).Error
	require.NoError(t, err)

	return org
}

// createDepartmentTestLocation создает тестовую локацию
func createDepartmentTestLocation(t *testing.T, db *gorm.DB, orgID int64, name string) *models.Location {
	location := &models.Location{
		OrganizationID: orgID,
		Name:           name,
		Source:         "manual",
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err)
	return location
}

// TestCreateDepartment тестирует создание отдела
func TestCreateDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	nonFounder := createDepartmentTestUser(t, db, "nonfounder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	tests := []struct {
		name        string
		locationID  int64
		userID      int64
		input       CreateDepartmentInput
		expectError bool
		errorCheck  func(*testing.T, error)
		checkDept   func(*testing.T, *models.Department)
	}{
		{
			name:       "Success - Create department",
			locationID: location.ID,
			userID:     founder.ID,
			input: CreateDepartmentInput{
				Name:        "IT Department",
				Description: stringPtr("Information Technology"),
			},
			expectError: false,
			checkDept: func(t *testing.T, dept *models.Department) {
				assert.Equal(t, "IT Department", dept.Name)
				assert.Equal(t, "Information Technology", *dept.Description)
				assert.Equal(t, location.ID, dept.LocationID)
				assert.Nil(t, dept.ParentID)
				assert.NotZero(t, dept.ID)
			},
		},
		{
			name:       "Success - Create department with parent",
			locationID: location.ID,
			userID:     founder.ID,
			input: CreateDepartmentInput{
				Name:        "Sub Department",
				ParentID:    nil, // Будет установлен в тесте
				Description: stringPtr("Sub department"),
			},
			expectError: false,
			checkDept: func(t *testing.T, dept *models.Department) {
				assert.Equal(t, "Sub Department", dept.Name)
				assert.NotNil(t, dept.ParentID)
			},
		},
		{
			name:       "Error - Non-founder cannot create",
			locationID: location.ID,
			userID:     nonFounder.ID,
			input: CreateDepartmentInput{
				Name: "Unauthorized Department",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrUnauthorized)
			},
		},
		{
			name:       "Error - Location not found",
			locationID: 99999,
			userID:     founder.ID,
			input: CreateDepartmentInput{
				Name: "Invalid Location Dept",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrLocationNotFound)
			},
		},
	}

	var parentDept *models.Department
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Для второго теста устанавливаем ParentID
			if i == 1 && parentDept != nil {
				tt.input.ParentID = &parentDept.ID
			}

			dept, err := departmentService.CreateDepartment(tt.locationID, tt.userID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, dept)
				if tt.checkDept != nil {
					tt.checkDept(t, dept)
				}
				// Сохраняем первый департамент для использования в качестве родителя
				if i == 0 {
					parentDept = dept
				}
			}
		})
	}
}

// TestGetDepartment тестирует получение отдела
func TestGetDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	nonFounder := createDepartmentTestUser(t, db, "nonfounder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	// Создаем отдел
	department := &models.Department{
		LocationID:  location.ID,
		Name:        "Test Department",
		Description: stringPtr("Test Description"),
	}
	err := db.Create(department).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		deptID      int64
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name:        "Success - Get department by founder",
			deptID:      department.ID,
			userID:      founder.ID,
			expectError: false,
		},
		{
			name:        "Error - Non-founder cannot get",
			deptID:      department.ID,
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrUnauthorized)
			},
		},
		{
			name:        "Error - Department not found",
			deptID:      99999,
			userID:      founder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrDepartmentNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept, err := departmentService.GetDepartment(tt.deptID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, dept)
				assert.Equal(t, department.ID, dept.ID)
				assert.Equal(t, "Test Department", dept.Name)
			}
		})
	}
}

// TestGetLocationDepartments тестирует получение всех отделов локации
func TestGetLocationDepartments(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	// Создаем несколько отделов
	departments := []models.Department{
		{LocationID: location.ID, Name: "IT"},
		{LocationID: location.ID, Name: "HR"},
		{LocationID: location.ID, Name: "Finance"},
	}
	for _, dept := range departments {
		err := db.Create(&dept).Error
		require.NoError(t, err)
	}

	depts, err := departmentService.GetLocationDepartments(location.ID, founder.ID)
	require.NoError(t, err)
	assert.Len(t, depts, 3)
}

// TestUpdateDepartment тестирует обновление отдела
func TestUpdateDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	department := &models.Department{
		LocationID:  location.ID,
		Name:        "Old Name",
		Description: stringPtr("Old Description"),
	}
	err := db.Create(department).Error
	require.NoError(t, err)

	input := UpdateDepartmentInput{
		Name:        "New Name",
		Description: stringPtr("New Description"),
	}

	updated, err := departmentService.UpdateDepartment(department.ID, founder.ID, input)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "New Description", *updated.Description)
}

// TestDeleteDepartment тестирует удаление отдела
func TestDeleteDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	department := &models.Department{
		LocationID: location.ID,
		Name:       "To Delete",
	}
	err := db.Create(department).Error
	require.NoError(t, err)

	err = departmentService.DeleteDepartment(department.ID, founder.ID)
	require.NoError(t, err)

	// Проверяем, что отдел удален
	var deleted models.Department
	err = db.First(&deleted, department.ID).Error
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestDeleteDepartmentWithChildren тестирует, что нельзя удалить отдел с дочерними отделами
func TestDeleteDepartmentWithChildren(t *testing.T) {
	db := setupDepartmentTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	orgService := NewOrganizationService(orgRepo)
	locationService := NewLocationService(locationRepo, orgService)
	departmentService := NewDepartmentService(departmentRepo, locationService)

	founder := createDepartmentTestUser(t, db, "founder@example.com")
	org := createDepartmentTestOrg(t, db, founder.ID, "Test Org")
	location := createDepartmentTestLocation(t, db, org.ID, "Main Office")

	// Создаем родительский отдел
	parent := &models.Department{
		LocationID: location.ID,
		Name:       "Parent",
	}
	err := db.Create(parent).Error
	require.NoError(t, err)

	// Создаем дочерний отдел
	child := &models.Department{
		LocationID: location.ID,
		ParentID:   &parent.ID,
		Name:       "Child",
	}
	err = db.Create(child).Error
	require.NoError(t, err)

	// Пытаемся удалить родительский отдел
	err = departmentService.DeleteDepartment(parent.ID, founder.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete department with child departments")
}

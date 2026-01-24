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

// setupDepartmentTestDB создает тестовую базу данных в памяти
func setupDepartmentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	// Миграции
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.Location{},
		&models.Department{},
		&models.Position{},
		&models.Employee{},
		&models.Permission{},
	)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// createDeptTestUser создает тестового пользователя
func createDeptTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
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

// createDeptTestOrg создает тестовую организацию
func createDeptTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
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

// createDeptTestLocation создает тестовую локацию
func createDeptTestLocation(t *testing.T, db *gorm.DB, orgID int64, name string) *models.Location {
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

// TestCreateDepartment тестирует создание отдела
func TestCreateDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	tests := []struct {
		name        string
		locationID  int64
		input       dto.CreateDepartmentRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkDept   func(*testing.T, *models.Department)
	}{
		{
			name:       "Success - Create department",
			locationID: location.ID,
			input: dto.CreateDepartmentRequest{
				Name:        "Engineering",
				Description: stringPtr("Software development team"),
			},
			expectError: false,
			checkDept: func(t *testing.T, dept *models.Department) {
				assert.Equal(t, "Engineering", dept.Name)
				assert.Equal(t, "Software development team", *dept.Description)
				assert.Equal(t, location.ID, dept.LocationID)
				assert.Nil(t, dept.ParentID)
				assert.NotZero(t, dept.ID)
			},
		},
		{
			name:       "Error - Location not found",
			locationID: 99999,
			input: dto.CreateDepartmentRequest{
				Name: "Orphan Department",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrLocationNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept, err := departmentService.CreateDepartment(tt.locationID, tt.input)

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
			}
		})
	}
}

// TestCreateDepartmentWithParent тестирует создание отдела с родителем
func TestCreateDepartmentWithParent(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	// Создаем родительский отдел
	parentDept := &models.Department{
		LocationID:  location.ID,
		Name:        "Parent Department",
		Description: stringPtr("Parent"),
	}
	err := db.Create(parentDept).Error
	require.NoError(t, err)

	// Создаем другую локацию для теста
	otherLocation := createDeptTestLocation(t, db, org.ID, "Other Office")
	otherDept := &models.Department{
		LocationID:  otherLocation.ID,
		Name:        "Other Parent",
		Description: stringPtr("In other location"),
	}
	err = db.Create(otherDept).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		locationID  int64
		input       dto.CreateDepartmentRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkDept   func(*testing.T, *models.Department)
	}{
		{
			name:       "Success - Create child department",
			locationID: location.ID,
			input: dto.CreateDepartmentRequest{
				Name:     "Child Department",
				ParentID: &parentDept.ID,
			},
			expectError: false,
			checkDept: func(t *testing.T, dept *models.Department) {
				assert.Equal(t, "Child Department", dept.Name)
				assert.Equal(t, parentDept.ID, *dept.ParentID)
				assert.Equal(t, location.ID, dept.LocationID)
			},
		},
		{
			name:       "Error - Parent not found",
			locationID: location.ID,
			input: dto.CreateDepartmentRequest{
				Name:     "Orphan Child",
				ParentID: int64Ptr(99999),
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "parent department not found")
			},
		},
		{
			name:       "Error - Parent in different location",
			locationID: location.ID,
			input: dto.CreateDepartmentRequest{
				Name:     "Cross Location Child",
				ParentID: &otherDept.ID,
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "parent department belongs to different location")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dept, err := departmentService.CreateDepartment(tt.locationID, tt.input)

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
			}
		})
	}
}

// TestGetDepartment тестирует получение отдела по ID
func TestGetDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	// Создаем отдел
	dept := &models.Department{
		LocationID:  location.ID,
		Name:        "Test Department",
		Description: stringPtr("Test description"),
	}
	err := db.Create(dept).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		deptID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
		checkDept   func(*testing.T, *models.Department)
	}{
		{
			name:        "Success - Get department",
			deptID:      dept.ID,
			expectError: false,
			checkDept: func(t *testing.T, d *models.Department) {
				assert.Equal(t, dept.ID, d.ID)
				assert.Equal(t, "Test Department", d.Name)
			},
		},
		{
			name:        "Error - Department not found",
			deptID:      99999,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrDepartmentNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := departmentService.GetDepartment(tt.deptID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, d)
				if tt.checkDept != nil {
					tt.checkDept(t, d)
				}
			}
		})
	}
}

// TestGetLocationDepartments тестирует получение всех отделов локации
func TestGetLocationDepartments(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	// Создаем несколько отделов
	depts := []models.Department{
		{LocationID: location.ID, Name: "Department 1"},
		{LocationID: location.ID, Name: "Department 2"},
		{LocationID: location.ID, Name: "Department 3"},
	}
	for i := range depts {
		err := db.Create(&depts[i]).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		locationID  int64
		expectError bool
		checkDepts  func(*testing.T, []models.Department)
	}{
		{
			name:        "Success - Get all departments",
			locationID:  location.ID,
			expectError: false,
			checkDepts: func(t *testing.T, depts []models.Department) {
				assert.Len(t, depts, 3)
			},
		},
		{
			name:        "Success - Empty list for non-existent location",
			locationID:  99999,
			expectError: false,
			checkDepts: func(t *testing.T, depts []models.Department) {
				assert.Len(t, depts, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depts, err := departmentService.GetLocationDepartments(tt.locationID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkDepts != nil {
					tt.checkDepts(t, depts)
				}
			}
		})
	}
}

// TestUpdateDepartment тестирует обновление отдела
func TestUpdateDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	// Создаем отдел
	dept := &models.Department{
		LocationID:  location.ID,
		Name:        "Old Name",
		Description: stringPtr("Old description"),
	}
	err := db.Create(dept).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		deptID      int64
		input       dto.UpdateDepartmentRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkDept   func(*testing.T, *models.Department)
	}{
		{
			name:   "Success - Update name",
			deptID: dept.ID,
			input: dto.UpdateDepartmentRequest{
				Name: "New Name",
			},
			expectError: false,
			checkDept: func(t *testing.T, d *models.Department) {
				assert.Equal(t, "New Name", d.Name)
			},
		},
		{
			name:   "Error - Department not found",
			deptID: 99999,
			input: dto.UpdateDepartmentRequest{
				Name: "Nonexistent",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrDepartmentNotFound)
			},
		},
		{
			name:   "Error - Cannot be its own parent",
			deptID: dept.ID,
			input: dto.UpdateDepartmentRequest{
				ParentID: &dept.ID,
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "cannot be its own parent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := departmentService.UpdateDepartment(tt.deptID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, d)
				if tt.checkDept != nil {
					tt.checkDept(t, d)
				}
			}
		})
	}
}

// TestDeleteDepartment тестирует удаление отдела
func TestDeleteDepartment(t *testing.T) {
	db := setupDepartmentTestDB(t)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)

	departmentService := NewDepartmentService(departmentRepo, locationRepo)

	founder := createDeptTestUser(t, db, "founder@example.com")
	org := createDeptTestOrg(t, db, founder.ID, "Test Org")
	location := createDeptTestLocation(t, db, org.ID, "Main Office")

	tests := []struct {
		name        string
		setupDept   func() int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "Success - Delete department",
			setupDept: func() int64 {
				dept := &models.Department{
					LocationID: location.ID,
					Name:       "To Delete",
				}
				err := db.Create(dept).Error
				require.NoError(t, err)
				return dept.ID
			},
			expectError: false,
		},
		{
			name: "Error - Department not found",
			setupDept: func() int64 {
				return 99999
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrDepartmentNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deptID := tt.setupDept()
			err := departmentService.DeleteDepartment(deptID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)

				// Проверяем, что отдел удален
				var deleted models.Department
				err := db.First(&deleted, deptID).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

// int64Ptr возвращает указатель на int64
func int64Ptr(i int64) *int64 {
	return &i
}

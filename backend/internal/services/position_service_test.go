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

// setupPositionTestDB создает тестовую базу данных в памяти для позиций
func setupPositionTestDB(t *testing.T) *gorm.DB {
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
		&models.Position{},
		&models.Employee{},
		&models.Permission{},
	)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// createPosTestUser создает тестового пользователя
func createPosTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
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

// createPosTestOrg создает тестовую организацию
func createPosTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
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

// createPosTestDepartment создает тестовый отдел
func createPosTestDepartment(t *testing.T, db *gorm.DB, locationID int64, name string) *models.Department {
	dept := &models.Department{
		LocationID: locationID,
		Name:       name,
	}
	err := db.Create(dept).Error
	require.NoError(t, err, "failed to create test department")
	return dept
}

// TestCreatePosition тестирует создание позиции
func TestCreatePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	// Создаем локацию и отдел для тестов с department_id
	location := &models.Location{
		OrganizationID: org.ID,
		Name:           "Main Office",
		Source:         "manual",
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err)

	dept := createPosTestDepartment(t, db, location.ID, "Engineering")

	tests := []struct {
		name        string
		orgID       int64
		input       dto.CreatePositionRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkPos    func(*testing.T, *models.Position)
	}{
		{
			name:  "Success - Create position without department",
			orgID: org.ID,
			input: dto.CreatePositionRequest{
				Name:        "Software Engineer",
				Description: stringPtr("Develop software"),
				IsAdmin:     false,
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Software Engineer", pos.Name)
				assert.Equal(t, "Develop software", *pos.Description)
				assert.False(t, pos.IsAdmin)
				assert.Nil(t, pos.DepartmentID)
				assert.Equal(t, org.ID, pos.OrganizationID)
				assert.NotZero(t, pos.ID)
			},
		},
		{
			name:  "Success - Create position with department",
			orgID: org.ID,
			input: dto.CreatePositionRequest{
				Name:         "Team Lead",
				DepartmentID: &dept.ID,
				IsAdmin:      true,
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Team Lead", pos.Name)
				assert.Equal(t, dept.ID, *pos.DepartmentID)
				assert.True(t, pos.IsAdmin)
			},
		},
		{
			name:  "Success - Create admin position",
			orgID: org.ID,
			input: dto.CreatePositionRequest{
				Name:    "CTO",
				IsAdmin: true,
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "CTO", pos.Name)
				assert.True(t, pos.IsAdmin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := positionService.CreatePosition(tt.orgID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, pos)
				if tt.checkPos != nil {
					tt.checkPos(t, pos)
				}
			}
		})
	}
}

// TestGetPosition тестирует получение позиции по ID
func TestGetPosition(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	// Создаем позицию
	position := &models.Position{
		OrganizationID: org.ID,
		Name:           "Test Position",
		Description:    stringPtr("Test description"),
		IsAdmin:        false,
	}
	err := db.Create(position).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		posID       int64
		expectError bool
		errorCheck  func(*testing.T, error)
		checkPos    func(*testing.T, *models.Position)
	}{
		{
			name:        "Success - Get position",
			posID:       position.ID,
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, position.ID, pos.ID)
				assert.Equal(t, "Test Position", pos.Name)
				assert.Equal(t, org.ID, pos.OrganizationID)
			},
		},
		{
			name:        "Error - Position not found",
			posID:       99999,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrPositionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := positionService.GetPosition(tt.posID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, pos)
				if tt.checkPos != nil {
					tt.checkPos(t, pos)
				}
			}
		})
	}
}

// TestGetOrganizationPositions тестирует получение всех позиций организации
func TestGetOrganizationPositions(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	// Создаем несколько позиций
	positions := []models.Position{
		{OrganizationID: org.ID, Name: "Position 1"},
		{OrganizationID: org.ID, Name: "Position 2"},
		{OrganizationID: org.ID, Name: "Position 3"},
	}
	for i := range positions {
		err := db.Create(&positions[i]).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		orgID       int64
		expectError bool
		checkPos    func(*testing.T, []models.Position)
	}{
		{
			name:        "Success - Get all positions",
			orgID:       org.ID,
			expectError: false,
			checkPos: func(t *testing.T, positions []models.Position) {
				assert.Len(t, positions, 3)
			},
		},
		{
			name:        "Success - Empty list for non-existent org",
			orgID:       99999,
			expectError: false,
			checkPos: func(t *testing.T, positions []models.Position) {
				assert.Len(t, positions, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positions, err := positionService.GetOrganizationPositions(tt.orgID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkPos != nil {
					tt.checkPos(t, positions)
				}
			}
		})
	}
}

// TestGetDepartmentPositions тестирует получение всех позиций отдела
func TestGetDepartmentPositions(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	// Создаем локацию и отдел
	location := &models.Location{
		OrganizationID: org.ID,
		Name:           "Main Office",
		Source:         "manual",
		IsActive:       true,
	}
	err := db.Create(location).Error
	require.NoError(t, err)

	dept := createPosTestDepartment(t, db, location.ID, "Engineering")

	// Создаем позиции для отдела
	positions := []models.Position{
		{OrganizationID: org.ID, DepartmentID: &dept.ID, Name: "Engineer"},
		{OrganizationID: org.ID, DepartmentID: &dept.ID, Name: "Senior Engineer"},
	}
	for i := range positions {
		err := db.Create(&positions[i]).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		deptID      int64
		expectError bool
		checkPos    func(*testing.T, []models.Position)
	}{
		{
			name:        "Success - Get department positions",
			deptID:      dept.ID,
			expectError: false,
			checkPos: func(t *testing.T, positions []models.Position) {
				assert.Len(t, positions, 2)
			},
		},
		{
			name:        "Success - Empty list for non-existent department",
			deptID:      99999,
			expectError: false,
			checkPos: func(t *testing.T, positions []models.Position) {
				assert.Len(t, positions, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positions, err := positionService.GetDepartmentPositions(tt.deptID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkPos != nil {
					tt.checkPos(t, positions)
				}
			}
		})
	}
}

// TestUpdatePosition тестирует обновление позиции
func TestUpdatePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	// Создаем позицию
	position := &models.Position{
		OrganizationID: org.ID,
		Name:           "Old Name",
		Description:    stringPtr("Old description"),
		IsAdmin:        false,
	}
	err := db.Create(position).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		posID       int64
		input       dto.UpdatePositionRequest
		expectError bool
		errorCheck  func(*testing.T, error)
		checkPos    func(*testing.T, *models.Position)
	}{
		{
			name:  "Success - Update name",
			posID: position.ID,
			input: dto.UpdatePositionRequest{
				Name: "New Name",
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "New Name", pos.Name)
			},
		},
		{
			name:  "Success - Update isAdmin",
			posID: position.ID,
			input: dto.UpdatePositionRequest{
				IsAdmin: boolPtr(true),
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.True(t, pos.IsAdmin)
			},
		},
		{
			name:  "Error - Position not found",
			posID: 99999,
			input: dto.UpdatePositionRequest{
				Name: "Nonexistent",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrPositionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := positionService.UpdatePosition(tt.posID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, pos)
				if tt.checkPos != nil {
					tt.checkPos(t, pos)
				}
			}
		})
	}
}

// TestDeletePosition тестирует удаление позиции
func TestDeletePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	positionRepo := repository.NewPositionRepository(db)

	positionService := NewPositionService(positionRepo)

	founder := createPosTestUser(t, db, "founder@example.com")
	org := createPosTestOrg(t, db, founder.ID, "Test Org")

	tests := []struct {
		name        string
		setupPos    func() int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "Success - Delete position",
			setupPos: func() int64 {
				pos := &models.Position{
					OrganizationID: org.ID,
					Name:           "To Delete",
				}
				err := db.Create(pos).Error
				require.NoError(t, err)
				return pos.ID
			},
			expectError: false,
		},
		{
			name: "Error - Position not found",
			setupPos: func() int64 {
				return 99999
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrPositionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posID := tt.setupPos()
			err := positionService.DeletePosition(posID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)

				// Проверяем, что позиция удалена
				var deleted models.Position
				err := db.First(&deleted, posID).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

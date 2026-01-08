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

// setupPositionTestDB создает тестовую базу данных для позиций
func setupPositionTestDB(t *testing.T) *gorm.DB {
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

// createPositionTestUser создает тестового пользователя
func createPositionTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
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

// createPositionTestOrg создает тестовую организацию
func createPositionTestOrg(t *testing.T, db *gorm.DB, userID int64, name string) *models.Organization {
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

// createPositionTestLocation создает тестовую локацию
func createPositionTestLocation(t *testing.T, db *gorm.DB, orgID int64, name string) *models.Location {
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

// createPositionTestDepartment создает тестовый отдел
func createPositionTestDepartment(t *testing.T, db *gorm.DB, locationID int64, name string) *models.Department {
	department := &models.Department{
		LocationID: locationID,
		Name:       name,
	}
	err := db.Create(department).Error
	require.NoError(t, err)
	return department
}

// TestCreatePosition тестирует создание позиции
func TestCreatePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	nonFounder := createPositionTestUser(t, db, "nonfounder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")
	location := createPositionTestLocation(t, db, org.ID, "Main Office")
	department := createPositionTestDepartment(t, db, location.ID, "IT Department")

	tests := []struct {
		name        string
		orgID       int64
		userID      int64
		input       CreatePositionInput
		expectError bool
		errorCheck  func(*testing.T, error)
		checkPos    func(*testing.T, *models.Position)
	}{
		{
			name:   "Success - Create position without department",
			orgID:  org.ID,
			userID: founder.ID,
			input: CreatePositionInput{
				Name:        "Software Engineer",
				IsAdmin:     false,
				Description: stringPtr("Develops software"),
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Software Engineer", pos.Name)
				assert.False(t, pos.IsAdmin)
				assert.Equal(t, "Develops software", *pos.Description)
				assert.Nil(t, pos.DepartmentID)
				assert.Equal(t, org.ID, pos.OrganizationID)
			},
		},
		{
			name:   "Success - Create position with department",
			orgID:  org.ID,
			userID: founder.ID,
			input: CreatePositionInput{
				Name:         "Senior Developer",
				DepartmentID: &department.ID,
				IsAdmin:      false,
				Description:  stringPtr("Senior level developer"),
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Senior Developer", pos.Name)
				assert.NotNil(t, pos.DepartmentID)
				assert.Equal(t, department.ID, *pos.DepartmentID)
			},
		},
		{
			name:   "Success - Create admin position",
			orgID:  org.ID,
			userID: founder.ID,
			input: CreatePositionInput{
				Name:    "Administrator",
				IsAdmin: true,
			},
			expectError: false,
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Administrator", pos.Name)
				assert.True(t, pos.IsAdmin)
			},
		},
		{
			name:   "Error - Non-founder cannot create",
			orgID:  org.ID,
			userID: nonFounder.ID,
			input: CreatePositionInput{
				Name: "Unauthorized Position",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrUnauthorized)
			},
		},
		{
			name:   "Error - Organization not found",
			orgID:  99999,
			userID: founder.ID,
			input: CreatePositionInput{
				Name: "Invalid Org Position",
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrOrganizationNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := positionService.CreatePosition(tt.orgID, tt.userID, tt.input)

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

// TestGetPosition тестирует получение позиции
func TestGetPosition(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	nonFounder := createPositionTestUser(t, db, "nonfounder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")

	// Создаем позицию
	position := &models.Position{
		OrganizationID: org.ID,
		Name:           "Test Position",
		IsAdmin:        false,
		Description:    stringPtr("Test Description"),
	}
	err := db.Create(position).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		posID       int64
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name:        "Success - Get position by founder",
			posID:       position.ID,
			userID:      founder.ID,
			expectError: false,
		},
		{
			name:        "Error - Non-founder cannot get",
			posID:       position.ID,
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrUnauthorized)
			},
		},
		{
			name:        "Error - Position not found",
			posID:       99999,
			userID:      founder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrPositionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := positionService.GetPosition(tt.posID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, pos)
				assert.Equal(t, position.ID, pos.ID)
				assert.Equal(t, "Test Position", pos.Name)
			}
		})
	}
}

// TestGetOrganizationPositions тестирует получение всех позиций организации
func TestGetOrganizationPositions(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")

	// Создаем несколько позиций
	positions := []models.Position{
		{OrganizationID: org.ID, Name: "Developer", IsAdmin: false},
		{OrganizationID: org.ID, Name: "Manager", IsAdmin: false},
		{OrganizationID: org.ID, Name: "Admin", IsAdmin: true},
	}
	for _, pos := range positions {
		err := db.Create(&pos).Error
		require.NoError(t, err)
	}

	poss, err := positionService.GetOrganizationPositions(org.ID, founder.ID)
	require.NoError(t, err)
	assert.Len(t, poss, 3)

	// Проверяем, что есть админская позиция
	hasAdmin := false
	for _, p := range poss {
		if p.IsAdmin {
			hasAdmin = true
			break
		}
	}
	assert.True(t, hasAdmin)
}

// TestGetDepartmentPositions тестирует получение всех позиций отдела
func TestGetDepartmentPositions(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")
	location := createPositionTestLocation(t, db, org.ID, "Main Office")
	department := createPositionTestDepartment(t, db, location.ID, "IT Department")

	// Создаем позиции в отделе
	positions := []models.Position{
		{OrganizationID: org.ID, DepartmentID: &department.ID, Name: "Junior Dev"},
		{OrganizationID: org.ID, DepartmentID: &department.ID, Name: "Senior Dev"},
	}
	for _, pos := range positions {
		err := db.Create(&pos).Error
		require.NoError(t, err)
	}

	// Создаем позицию без отдела (не должна попасть в результат)
	otherPos := &models.Position{
		OrganizationID: org.ID,
		Name:           "Other Position",
	}
	err := db.Create(otherPos).Error
	require.NoError(t, err)

	poss, err := positionService.GetDepartmentPositions(department.ID, founder.ID)
	require.NoError(t, err)
	assert.Len(t, poss, 2)
}

// TestUpdatePosition тестирует обновление позиции
func TestUpdatePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")

	position := &models.Position{
		OrganizationID: org.ID,
		Name:           "Old Name",
		IsAdmin:        false,
		Description:    stringPtr("Old Description"),
	}
	err := db.Create(position).Error
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    UpdatePositionInput
		checkPos func(*testing.T, *models.Position)
	}{
		{
			name: "Update all fields",
			input: UpdatePositionInput{
				Name:        "New Name",
				IsAdmin:     boolPtr(true),
				Description: stringPtr("New Description"),
			},
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "New Name", pos.Name)
				assert.True(t, pos.IsAdmin)
				assert.Equal(t, "New Description", *pos.Description)
			},
		},
		{
			name: "Partial update",
			input: UpdatePositionInput{
				Name: "Another Name",
			},
			checkPos: func(t *testing.T, pos *models.Position) {
				assert.Equal(t, "Another Name", pos.Name)
				// IsAdmin должен остаться true от предыдущего обновления
				assert.True(t, pos.IsAdmin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := positionService.UpdatePosition(position.ID, founder.ID, tt.input)
			require.NoError(t, err)
			if tt.checkPos != nil {
				tt.checkPos(t, updated)
			}
		})
	}
}

// TestDeletePosition тестирует удаление позиции
func TestDeletePosition(t *testing.T) {
	db := setupPositionTestDB(t)
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)

	orgService := NewOrganizationService(orgRepo)
	positionService := NewPositionService(positionRepo, orgService)

	founder := createPositionTestUser(t, db, "founder@example.com")
	nonFounder := createPositionTestUser(t, db, "nonfounder@example.com")
	org := createPositionTestOrg(t, db, founder.ID, "Test Org")

	tests := []struct {
		name        string
		setupPos    func() *models.Position
		userID      int64
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "Success - Delete position",
			setupPos: func() *models.Position {
				pos := &models.Position{
					OrganizationID: org.ID,
					Name:           "To Delete",
				}
				err := db.Create(pos).Error
				require.NoError(t, err)
				return pos
			},
			userID:      founder.ID,
			expectError: false,
		},
		{
			name: "Error - Non-founder cannot delete",
			setupPos: func() *models.Position {
				pos := &models.Position{
					OrganizationID: org.ID,
					Name:           "Protected",
				}
				err := db.Create(pos).Error
				require.NoError(t, err)
				return pos
			},
			userID:      nonFounder.ID,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrUnauthorized)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := tt.setupPos()
			err := positionService.DeletePosition(pos.ID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)

				// Проверяем, что позиция удалена
				var deleted models.Position
				err := db.First(&deleted, pos.ID).Error
				assert.Error(t, err)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

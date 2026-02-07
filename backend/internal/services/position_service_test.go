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

// =============================================================================
// DEPARTMENT SERVICE TESTS
// =============================================================================

type DepartmentServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	service      *DepartmentService
	testOrg      *models.Organization
	testLocation *models.Location
}

func (s *DepartmentServiceTestSuite) SetupTest() {
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
	)
	s.Require().NoError(err)

	locationRepo := repository.NewLocationRepository(s.db)
	deptRepo := repository.NewDepartmentRepository(s.db)
	s.service = NewDepartmentService(deptRepo, locationRepo)

	// Создаём тестовые данные
	s.testOrg = s.createTestOrg()
	s.testLocation = s.createTestLocation(s.testOrg.ID, "Main Office")
}

func (s *DepartmentServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *DepartmentServiceTestSuite) createTestOrg() *models.Organization {
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	s.Require().NoError(s.db.Create(org).Error)
	return org
}

func (s *DepartmentServiceTestSuite) createTestLocation(orgID int64, name string) *models.Location {
	loc := &models.Location{
		OrganizationID: orgID,
		Name:           name,
		Source:         "manual",
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(loc).Error)
	return loc
}

func (s *DepartmentServiceTestSuite) TestCreateDepartment_Success() {
	testCases := []struct {
		name      string
		setupData func() dto.CreateDepartmentRequest
		checkDept func(*models.Department)
	}{
		{
			name: "without parent",
			setupData: func() dto.CreateDepartmentRequest {
				return dto.CreateDepartmentRequest{
					Name:        "Engineering",
					Description: stringPtr("Software development team"),
				}
			},
			checkDept: func(dept *models.Department) {
				s.Assert().Equal("Engineering", dept.Name)
				s.Assert().Equal("Software development team", *dept.Description)
				s.Assert().Equal(s.testLocation.ID, dept.LocationID)
				s.Assert().Nil(dept.ParentID)
			},
		},
		{
			name: "with parent",
			setupData: func() dto.CreateDepartmentRequest {
				parent := &models.Department{
					LocationID: s.testLocation.ID,
					Name:       "Parent Department",
				}
				s.Require().NoError(s.db.Create(parent).Error)

				return dto.CreateDepartmentRequest{
					Name:     "Child Department",
					ParentID: &parent.ID,
				}
			},
			checkDept: func(dept *models.Department) {
				s.Assert().Equal("Child Department", dept.Name)
				s.Assert().NotNil(dept.ParentID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			input := tc.setupData()
			dept, err := s.service.CreateDepartment(s.testLocation.ID, input)
			s.Require().NoError(err)
			s.Assert().NotNil(dept)
			tc.checkDept(dept)
		})
	}
}

func (s *DepartmentServiceTestSuite) TestCreateDepartment_Validation() {
	// Создаём отдел в другой локации для теста
	otherLocation := s.createTestLocation(s.testOrg.ID, "Other Office")
	otherDept := &models.Department{
		LocationID: otherLocation.ID,
		Name:       "Other Department",
	}
	s.Require().NoError(s.db.Create(otherDept).Error)

	testCases := []struct {
		name        string
		locationID  int64
		input       dto.CreateDepartmentRequest
		expectErr   bool
		errContains string
	}{
		{
			name:        "location not found",
			locationID:  99999,
			input:       dto.CreateDepartmentRequest{Name: "Orphan"},
			expectErr:   true,
			errContains: "",
		},
		{
			name:       "parent not found",
			locationID: s.testLocation.ID,
			input: dto.CreateDepartmentRequest{
				Name:     "Child",
				ParentID: int64Ptr(99999),
			},
			expectErr:   true,
			errContains: "parent department not found",
		},
		{
			name:       "parent in different location",
			locationID: s.testLocation.ID,
			input: dto.CreateDepartmentRequest{
				Name:     "Cross Location Child",
				ParentID: &otherDept.ID,
			},
			expectErr:   true,
			errContains: "different location",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			dept, err := s.service.CreateDepartment(tc.locationID, tc.input)

			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Nil(dept)
				if tc.errContains != "" {
					s.Assert().Contains(err.Error(), tc.errContains)
				}
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

func (s *DepartmentServiceTestSuite) TestUpdateDepartment() {
	dept := &models.Department{
		LocationID:  s.testLocation.ID,
		Name:        "Old Name",
		Description: stringPtr("Old description"),
	}
	s.Require().NoError(s.db.Create(dept).Error)

	testCases := []struct {
		name      string
		input     dto.UpdateDepartmentRequest
		expectErr bool
		checkDept func(*models.Department)
	}{
		{
			name:      "update name",
			input:     dto.UpdateDepartmentRequest{Name: "New Name"},
			expectErr: false,
			checkDept: func(d *models.Department) {
				s.Assert().Equal("New Name", d.Name)
			},
		},
		{
			name:      "cannot be own parent",
			input:     dto.UpdateDepartmentRequest{ParentID: &dept.ID},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.service.UpdateDepartment(dept.ID, tc.input)

			if tc.expectErr {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
				tc.checkDept(result)
			}
		})
	}
}

func (s *DepartmentServiceTestSuite) TestDeleteDepartment() {
	dept := &models.Department{
		LocationID: s.testLocation.ID,
		Name:       "To Delete",
	}
	s.Require().NoError(s.db.Create(dept).Error)

	err := s.service.DeleteDepartment(dept.ID)
	s.Assert().NoError(err)

	// Проверяем удаление
	var deleted models.Department
	err = s.db.First(&deleted, dept.ID).Error
	s.Assert().ErrorIs(err, gorm.ErrRecordNotFound)
}

func TestDepartmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DepartmentServiceTestSuite))
}

// =============================================================================
// POSITION SERVICE TESTS
// =============================================================================

type PositionServiceTestSuite struct {
	suite.Suite
	db       *gorm.DB
	service  *PositionService
	testOrg  *models.Organization
	testDept *models.Department
}

func (s *PositionServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	err = s.db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Location{},
		&models.Department{},
		&models.Position{},
		&models.Employee{},
	)
	s.Require().NoError(err)

	posRepo := repository.NewPositionRepository(s.db)
	s.service = NewPositionService(posRepo)

	// Создаём тестовые данные
	s.testOrg = s.createTestOrg()
	location := s.createTestLocation(s.testOrg.ID)
	s.testDept = s.createTestDepartment(location.ID)
}

func (s *PositionServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *PositionServiceTestSuite) createTestOrg() *models.Organization {
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	s.Require().NoError(s.db.Create(org).Error)
	return org
}

func (s *PositionServiceTestSuite) createTestLocation(orgID int64) *models.Location {
	loc := &models.Location{
		OrganizationID: orgID,
		Name:           "Main Office",
		Source:         "manual",
		IsActive:       true,
	}
	s.Require().NoError(s.db.Create(loc).Error)
	return loc
}

func (s *PositionServiceTestSuite) createTestDepartment(locationID int64) *models.Department {
	dept := &models.Department{
		LocationID: locationID,
		Name:       "Engineering",
	}
	s.Require().NoError(s.db.Create(dept).Error)
	return dept
}

func (s *PositionServiceTestSuite) TestCreatePosition_Success() {
	testCases := []struct {
		name     string
		input    dto.CreatePositionRequest
		checkPos func(*models.Position)
	}{
		{
			name: "without department",
			input: dto.CreatePositionRequest{
				Name:        "Software Engineer",
				Description: stringPtr("Develop software"),
				IsAdmin:     false,
			},
			checkPos: func(pos *models.Position) {
				s.Assert().Equal("Software Engineer", pos.Name)
				s.Assert().Equal("Develop software", *pos.Description)
				s.Assert().False(pos.IsAdmin)
				s.Assert().Nil(pos.DepartmentID)
			},
		},
		{
			name: "with department",
			input: dto.CreatePositionRequest{
				Name:         "Team Lead",
				DepartmentID: &s.testDept.ID,
				IsAdmin:      true,
			},
			checkPos: func(pos *models.Position) {
				s.Assert().Equal("Team Lead", pos.Name)
				s.Assert().Equal(s.testDept.ID, *pos.DepartmentID)
				s.Assert().True(pos.IsAdmin)
			},
		},
		{
			name: "admin position",
			input: dto.CreatePositionRequest{
				Name:    "CTO",
				IsAdmin: true,
			},
			checkPos: func(pos *models.Position) {
				s.Assert().True(pos.IsAdmin)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			pos, err := s.service.CreatePosition(s.testOrg.ID, tc.input)
			s.Require().NoError(err)
			s.Assert().NotNil(pos)
			tc.checkPos(pos)
		})
	}
}

func (s *PositionServiceTestSuite) TestGetOrganizationPositions() {
	// Создаём несколько позиций
	positions := []models.Position{
		{OrganizationID: s.testOrg.ID, Name: "Position 1"},
		{OrganizationID: s.testOrg.ID, Name: "Position 2"},
		{OrganizationID: s.testOrg.ID, Name: "Position 3"},
	}
	for i := range positions {
		s.Require().NoError(s.db.Create(&positions[i]).Error)
	}

	// Получаем все позиции
	result, err := s.service.GetOrganizationPositions(s.testOrg.ID)
	s.Assert().NoError(err)
	s.Assert().Len(result, 3)
}

func (s *PositionServiceTestSuite) TestGetDepartmentPositions() {
	// Создаём позиции для отдела
	positions := []models.Position{
		{OrganizationID: s.testOrg.ID, DepartmentID: &s.testDept.ID, Name: "Engineer"},
		{OrganizationID: s.testOrg.ID, DepartmentID: &s.testDept.ID, Name: "Senior Engineer"},
	}
	for i := range positions {
		s.Require().NoError(s.db.Create(&positions[i]).Error)
	}

	// Получаем позиции отдела
	result, err := s.service.GetDepartmentPositions(s.testDept.ID)
	s.Assert().NoError(err)
	s.Assert().Len(result, 2)
}

func (s *PositionServiceTestSuite) TestUpdatePosition() {
	pos := &models.Position{
		OrganizationID: s.testOrg.ID,
		Name:           "Old Name",
		IsAdmin:        false,
	}
	s.Require().NoError(s.db.Create(pos).Error)

	testCases := []struct {
		name      string
		input     dto.UpdatePositionRequest
		expectErr bool
		checkPos  func(*models.Position)
	}{
		{
			name:      "update name",
			input:     dto.UpdatePositionRequest{Name: "New Name"},
			expectErr: false,
			checkPos: func(p *models.Position) {
				s.Assert().Equal("New Name", p.Name)
			},
		},
		{
			name:      "update isAdmin",
			input:     dto.UpdatePositionRequest{IsAdmin: boolPtr(true)},
			expectErr: false,
			checkPos: func(p *models.Position) {
				s.Assert().True(p.IsAdmin)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.service.UpdatePosition(pos.ID, tc.input)

			if tc.expectErr {
				s.Assert().Error(err)
			} else {
				s.Assert().NoError(err)
				tc.checkPos(result)
			}
		})
	}
}

func (s *PositionServiceTestSuite) TestDeletePosition() {
	pos := &models.Position{
		OrganizationID: s.testOrg.ID,
		Name:           "To Delete",
	}
	s.Require().NoError(s.db.Create(pos).Error)

	err := s.service.DeletePosition(pos.ID)
	s.Assert().NoError(err)

	// Проверяем удаление
	var deleted models.Position
	err = s.db.First(&deleted, pos.ID).Error
	s.Assert().ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *PositionServiceTestSuite) TestPositionNotFound() {
	_, err := s.service.GetPosition(99999)
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, apperrors.ErrPositionNotFound)
}

func TestPositionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PositionServiceTestSuite))
}

// Helper functions
func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

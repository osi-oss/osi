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

// setupPermissionsTestDB создает тестовую базу данных с таблицей прав
func setupPermissionsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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
	require.NoError(t, err)

	// Создаем junction таблицы вручную (AutoMigrate их не создает)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS member_permissions (
			organization_member_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			PRIMARY KEY (organization_organization_member_id, permission_id)
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS position_permissions (
			position_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			PRIMARY KEY (position_id, permission_id)
		)
	`).Error
	require.NoError(t, err)

	// Создаем тестовые права в БД
	permissions := []models.Permission{
		{Code: "departments.create", Description: "Create departments", GroupName: "Departments"},
		{Code: "positions.create", Description: "Create positions", GroupName: "Positions"},
		{Code: "members.invite", Description: "Invite members", GroupName: "Members"},
		{Code: "members.view", Description: "View members", GroupName: "Members"},
	}
	for _, perm := range permissions {
		err := db.Create(&perm).Error
		require.NoError(t, err)
	}

	return db
}

// TestLocationPermissions проверяет работу прав для операций с локациями
func TestLocationPermissions(t *testing.T) {
	db := setupPermissionsTestDB(t)

	// Репозитории
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	// Сервисы
	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	orgService.SetPermissionService(permissionService)
	locationService := NewLocationService(locationRepo, orgService, permissionService)

	// Создаем пользователей
	founder := &models.User{Email: stringPtr("founder@test.com"), PasswordHash: "hash"}
	memberWithPermission := &models.User{Email: stringPtr("member_with@test.com"), PasswordHash: "hash"}
	memberWithoutPermission := &models.User{Email: stringPtr("member_without@test.com"), PasswordHash: "hash"}
	require.NoError(t, db.Create(founder).Error)
	require.NoError(t, db.Create(memberWithPermission).Error)
	require.NoError(t, db.Create(memberWithoutPermission).Error)

	// Создаем организацию
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	require.NoError(t, db.Create(org).Error)

	// Добавляем founder
	orgFounder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         founder.ID,
		IsMain:         true,
	}
	require.NoError(t, db.Create(orgFounder).Error)

	// Добавляем члена с правами
	memberWith := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWith).Error)

	// Добавляем члена без прав
	memberWithout := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithoutPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWithout).Error)

	// Получаем право departments.create
	var deptPermission models.Permission
	require.NoError(t, db.Where("code = ?", "departments.create").First(&deptPermission).Error)

	// Назначаем право члену с правами
	require.NoError(t, db.Exec("INSERT INTO member_permissions (organization_member_id, permission_id) VALUES (?, ?)",
		memberWith.ID, deptPermission.ID).Error)

	t.Run("Founder can create location", func(t *testing.T) {
		location, err := locationService.CreateLocation(org.ID, founder.ID, CreateLocationInput{
			Name:   "Founder Location",
			Source: "manual",
		})
		require.NoError(t, err)
		assert.NotNil(t, location)
		assert.Equal(t, "Founder Location", location.Name)
	})

	t.Run("Member with permission can create location", func(t *testing.T) {
		location, err := locationService.CreateLocation(org.ID, memberWithPermission.ID, CreateLocationInput{
			Name:   "Member Location",
			Source: "manual",
		})
		require.NoError(t, err)
		assert.NotNil(t, location)
		assert.Equal(t, "Member Location", location.Name)
	})

	t.Run("Member without permission cannot create location", func(t *testing.T) {
		location, err := locationService.CreateLocation(org.ID, memberWithoutPermission.ID, CreateLocationInput{
			Name:   "Unauthorized Location",
			Source: "manual",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, location)
	})

	t.Run("Member with permission can update location", func(t *testing.T) {
		// Создаем локацию
		location := &models.Location{
			OrganizationID: org.ID,
			Name:           "Test Location",
			Source:         "manual",
			IsActive:       true,
		}
		require.NoError(t, db.Create(location).Error)

		// Обновляем
		updated, err := locationService.UpdateLocation(location.ID, memberWithPermission.ID, UpdateLocationInput{
			Name: "Updated by Member",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated by Member", updated.Name)
	})

	t.Run("Member without permission cannot update location", func(t *testing.T) {
		// Создаем локацию
		location := &models.Location{
			OrganizationID: org.ID,
			Name:           "Test Location 2",
			Source:         "manual",
			IsActive:       true,
		}
		require.NoError(t, db.Create(location).Error)

		// Пытаемся обновить
		updated, err := locationService.UpdateLocation(location.ID, memberWithoutPermission.ID, UpdateLocationInput{
			Name: "Should Fail",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, updated)
	})

	t.Run("Member with permission can delete location", func(t *testing.T) {
		// Создаем локацию
		location := &models.Location{
			OrganizationID: org.ID,
			Name:           "To Delete",
			Source:         "manual",
			IsActive:       true,
		}
		require.NoError(t, db.Create(location).Error)

		// Удаляем
		err := locationService.DeleteLocation(location.ID, memberWithPermission.ID)
		require.NoError(t, err)
	})

	t.Run("Member without permission cannot delete location", func(t *testing.T) {
		// Создаем локацию
		location := &models.Location{
			OrganizationID: org.ID,
			Name:           "Cannot Delete",
			Source:         "manual",
			IsActive:       true,
		}
		require.NoError(t, db.Create(location).Error)

		// Пытаемся удалить
		err := locationService.DeleteLocation(location.ID, memberWithoutPermission.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("All members can view locations", func(t *testing.T) {
		// Founder может просматривать
		locations, err := locationService.GetOrganizationLocations(org.ID, founder.ID)
		require.NoError(t, err)
		assert.NotNil(t, locations)

		// Член с правами может просматривать
		locations, err = locationService.GetOrganizationLocations(org.ID, memberWithPermission.ID)
		require.NoError(t, err)
		assert.NotNil(t, locations)

		// Член без прав тоже может просматривать (не требует специального права)
		locations, err = locationService.GetOrganizationLocations(org.ID, memberWithoutPermission.ID)
		require.NoError(t, err)
		assert.NotNil(t, locations)
	})
}

// TestDepartmentPermissions проверяет работу прав для операций с отделами
func TestDepartmentPermissions(t *testing.T) {
	db := setupPermissionsTestDB(t)

	// Репозитории
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	// Сервисы
	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	orgService.SetPermissionService(permissionService)
	locationService := NewLocationService(locationRepo, orgService, permissionService)
	departmentService := NewDepartmentService(departmentRepo, locationService, permissionService)

	// Создаем пользователей
	founder := &models.User{Email: stringPtr("founder@test.com"), PasswordHash: "hash"}
	memberWithPermission := &models.User{Email: stringPtr("member_with@test.com"), PasswordHash: "hash"}
	memberWithoutPermission := &models.User{Email: stringPtr("member_without@test.com"), PasswordHash: "hash"}
	require.NoError(t, db.Create(founder).Error)
	require.NoError(t, db.Create(memberWithPermission).Error)
	require.NoError(t, db.Create(memberWithoutPermission).Error)

	// Создаем организацию
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	require.NoError(t, db.Create(org).Error)

	// Добавляем founder
	orgFounder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         founder.ID,
		IsMain:         true,
	}
	require.NoError(t, db.Create(orgFounder).Error)

	// Создаем локацию
	location := &models.Location{
		OrganizationID: org.ID,
		Name:           "Main Office",
		Source:         "manual",
		IsActive:       true,
	}
	require.NoError(t, db.Create(location).Error)

	// Добавляем члена с правами
	memberWith := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWith).Error)

	// Добавляем члена без прав
	memberWithout := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithoutPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWithout).Error)

	// Получаем право departments.create
	var deptPermission models.Permission
	require.NoError(t, db.Where("code = ?", "departments.create").First(&deptPermission).Error)

	// Назначаем право члену с правами
	require.NoError(t, db.Exec("INSERT INTO member_permissions (organization_member_id, permission_id) VALUES (?, ?)",
		memberWith.ID, deptPermission.ID).Error)

	t.Run("Founder can create department", func(t *testing.T) {
		dept, err := departmentService.CreateDepartment(location.ID, founder.ID, CreateDepartmentInput{
			Name: "Founder Department",
		})
		require.NoError(t, err)
		assert.NotNil(t, dept)
		assert.Equal(t, "Founder Department", dept.Name)
	})

	t.Run("Member with permission can create department", func(t *testing.T) {
		dept, err := departmentService.CreateDepartment(location.ID, memberWithPermission.ID, CreateDepartmentInput{
			Name: "Member Department",
		})
		require.NoError(t, err)
		assert.NotNil(t, dept)
		assert.Equal(t, "Member Department", dept.Name)
	})

	t.Run("Member without permission cannot create department", func(t *testing.T) {
		dept, err := departmentService.CreateDepartment(location.ID, memberWithoutPermission.ID, CreateDepartmentInput{
			Name: "Unauthorized Department",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, dept)
	})

	t.Run("Member with permission can update department", func(t *testing.T) {
		// Создаем отдел
		dept := &models.Department{
			LocationID: location.ID,
			Name:       "Test Department",
		}
		require.NoError(t, db.Create(dept).Error)

		// Обновляем
		updated, err := departmentService.UpdateDepartment(dept.ID, memberWithPermission.ID, UpdateDepartmentInput{
			Name: "Updated Department",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Department", updated.Name)
	})

	t.Run("Member without permission cannot update department", func(t *testing.T) {
		// Создаем отдел
		dept := &models.Department{
			LocationID: location.ID,
			Name:       "Test Department 2",
		}
		require.NoError(t, db.Create(dept).Error)

		// Пытаемся обновить
		updated, err := departmentService.UpdateDepartment(dept.ID, memberWithoutPermission.ID, UpdateDepartmentInput{
			Name: "Should Fail",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, updated)
	})

	t.Run("Member with permission can delete department", func(t *testing.T) {
		// Создаем отдел
		dept := &models.Department{
			LocationID: location.ID,
			Name:       "To Delete",
		}
		require.NoError(t, db.Create(dept).Error)

		// Удаляем
		err := departmentService.DeleteDepartment(dept.ID, memberWithPermission.ID)
		require.NoError(t, err)
	})

	t.Run("Member without permission cannot delete department", func(t *testing.T) {
		// Создаем отдел
		dept := &models.Department{
			LocationID: location.ID,
			Name:       "Cannot Delete",
		}
		require.NoError(t, db.Create(dept).Error)

		// Пытаемся удалить
		err := departmentService.DeleteDepartment(dept.ID, memberWithoutPermission.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})
}

// TestPositionPermissions проверяет работу прав для операций с позициями
func TestPositionPermissions(t *testing.T) {
	db := setupPermissionsTestDB(t)

	// Репозитории
	orgRepo := repository.NewOrganizationRepository(db)
	positionRepo := repository.NewPositionRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	// Сервисы
	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	orgService.SetPermissionService(permissionService)
	positionService := NewPositionService(positionRepo, orgService, permissionService)

	// Создаем пользователей
	founder := &models.User{Email: stringPtr("founder@test.com"), PasswordHash: "hash"}
	memberWithPermission := &models.User{Email: stringPtr("member_with@test.com"), PasswordHash: "hash"}
	memberWithoutPermission := &models.User{Email: stringPtr("member_without@test.com"), PasswordHash: "hash"}
	require.NoError(t, db.Create(founder).Error)
	require.NoError(t, db.Create(memberWithPermission).Error)
	require.NoError(t, db.Create(memberWithoutPermission).Error)

	// Создаем организацию
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	require.NoError(t, db.Create(org).Error)

	// Добавляем founder
	orgFounder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         founder.ID,
		IsMain:         true,
	}
	require.NoError(t, db.Create(orgFounder).Error)

	// Добавляем члена с правами
	memberWith := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWith).Error)

	// Добавляем члена без прав
	memberWithout := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         memberWithoutPermission.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(memberWithout).Error)

	// Получаем право positions.create
	var posPermission models.Permission
	require.NoError(t, db.Where("code = ?", "positions.create").First(&posPermission).Error)

	// Назначаем право члену с правами
	require.NoError(t, db.Exec("INSERT INTO member_permissions (organization_member_id, permission_id) VALUES (?, ?)",
		memberWith.ID, posPermission.ID).Error)

	t.Run("Founder can create position", func(t *testing.T) {
		pos, err := positionService.CreatePosition(org.ID, founder.ID, CreatePositionInput{
			Name: "Founder Position",
		})
		require.NoError(t, err)
		assert.NotNil(t, pos)
		assert.Equal(t, "Founder Position", pos.Name)
	})

	t.Run("Member with permission can create position", func(t *testing.T) {
		pos, err := positionService.CreatePosition(org.ID, memberWithPermission.ID, CreatePositionInput{
			Name: "Member Position",
		})
		require.NoError(t, err)
		assert.NotNil(t, pos)
		assert.Equal(t, "Member Position", pos.Name)
	})

	t.Run("Member without permission cannot create position", func(t *testing.T) {
		pos, err := positionService.CreatePosition(org.ID, memberWithoutPermission.ID, CreatePositionInput{
			Name: "Unauthorized Position",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, pos)
	})

	t.Run("Member with permission can update position", func(t *testing.T) {
		// Создаем позицию
		pos := &models.Position{
			OrganizationID: org.ID,
			Name:           "Test Position",
		}
		require.NoError(t, db.Create(pos).Error)

		// Обновляем
		updated, err := positionService.UpdatePosition(pos.ID, memberWithPermission.ID, UpdatePositionInput{
			Name: "Updated Position",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Position", updated.Name)
	})

	t.Run("Member without permission cannot update position", func(t *testing.T) {
		// Создаем позицию
		pos := &models.Position{
			OrganizationID: org.ID,
			Name:           "Test Position 2",
		}
		require.NoError(t, db.Create(pos).Error)

		// Пытаемся обновить
		updated, err := positionService.UpdatePosition(pos.ID, memberWithoutPermission.ID, UpdatePositionInput{
			Name: "Should Fail",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, updated)
	})

	t.Run("Member with permission can delete position", func(t *testing.T) {
		// Создаем позицию
		pos := &models.Position{
			OrganizationID: org.ID,
			Name:           "To Delete",
		}
		require.NoError(t, db.Create(pos).Error)

		// Удаляем
		err := positionService.DeletePosition(pos.ID, memberWithPermission.ID)
		require.NoError(t, err)
	})

	t.Run("Member without permission cannot delete position", func(t *testing.T) {
		// Создаем позицию
		pos := &models.Position{
			OrganizationID: org.ID,
			Name:           "Cannot Delete",
		}
		require.NoError(t, db.Create(pos).Error)

		// Пытаемся удалить
		err := positionService.DeletePosition(pos.ID, memberWithoutPermission.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("All members can view positions", func(t *testing.T) {
		// Founder может просматривать
		positions, err := positionService.GetOrganizationPositions(org.ID, founder.ID)
		require.NoError(t, err)
		assert.NotNil(t, positions)

		// Член с правами может просматривать
		positions, err = positionService.GetOrganizationPositions(org.ID, memberWithPermission.ID)
		require.NoError(t, err)
		assert.NotNil(t, positions)

		// Член без прав тоже может просматривать (не требует специального права)
		positions, err = positionService.GetOrganizationPositions(org.ID, memberWithoutPermission.ID)
		require.NoError(t, err)
		assert.NotNil(t, positions)
	})
}

// TestPermissionsThroughPosition проверяет наследование прав через позицию
func TestPermissionsThroughPosition(t *testing.T) {
	db := setupPermissionsTestDB(t)

	// Репозитории
	orgRepo := repository.NewOrganizationRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// Сервисы
	permissionService := NewPermissionService(permissionRepo, orgRepo, employeeRepo)
	orgService := NewOrganizationService(orgRepo)
	orgService.SetPermissionService(permissionService)
	locationService := NewLocationService(locationRepo, orgService, permissionService)

	// Создаем пользователей
	founder := &models.User{Email: stringPtr("founder@test.com"), PasswordHash: "hash"}
	employee := &models.User{Email: stringPtr("employee@test.com"), PasswordHash: "hash"}
	require.NoError(t, db.Create(founder).Error)
	require.NoError(t, db.Create(employee).Error)

	// Создаем организацию
	org := &models.Organization{Name: "Test Org", Status: models.OrgDraft}
	require.NoError(t, db.Create(org).Error)

	// Добавляем founder
	orgFounder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         founder.ID,
		IsMain:         true,
	}
	require.NoError(t, db.Create(orgFounder).Error)

	// Добавляем сотрудника как члена организации (без прямых прав)
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         employee.ID,
		Status:         models.MemberActive,
	}
	require.NoError(t, db.Create(member).Error)

	// Создаем позицию
	position := &models.Position{
		OrganizationID: org.ID,
		Name:           "Manager",
	}
	require.NoError(t, db.Create(position).Error)

	// Получаем право departments.create
	var deptPermission models.Permission
	require.NoError(t, db.Where("code = ?", "departments.create").First(&deptPermission).Error)

	// Назначаем право позиции
	require.NoError(t, db.Exec("INSERT INTO position_permissions (position_id, permission_id) VALUES (?, ?)",
		position.ID, deptPermission.ID).Error)

	t.Run("Employee without position cannot create location", func(t *testing.T) {
		// Сотрудник еще не назначен на позицию
		location, err := locationService.CreateLocation(org.ID, employee.ID, CreateLocationInput{
			Name:   "Should Fail",
			Source: "manual",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, location)
	})

	t.Run("Employee with position can create location", func(t *testing.T) {
		// Назначаем сотрудника на позицию
		employeeRecord := &models.Employee{
			MemberID:   member.ID,
			PositionID: position.ID,
		}
		require.NoError(t, db.Create(employeeRecord).Error)

		// Теперь сотрудник должен иметь право через позицию
		location, err := locationService.CreateLocation(org.ID, employee.ID, CreateLocationInput{
			Name:   "Employee Location",
			Source: "manual",
		})
		require.NoError(t, err)
		assert.NotNil(t, location)
		assert.Equal(t, "Employee Location", location.Name)
	})
}

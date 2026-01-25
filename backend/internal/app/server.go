package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/config"
	"github.com/osi-oss/osi/internal/controllers"
	"github.com/osi-oss/osi/internal/db"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/logger"
	"github.com/osi-oss/osi/internal/middleware"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/services"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Start(cfg *config.Config) {
	// Инициализируем логгер
	logger.Init(cfg.IsDev)
	logger.Info("🚀 Server starting", slog.String("port", cfg.AppPort))

	// Reading config && Connection to db
	dbConn, err := db.Connect(cfg)
	if err != nil {
		logger.Error("DB connection error", err, slog.String("host", cfg.PgHost), slog.String("db", cfg.PgDb))
		panic(err)
	}
	logger.Info("✅ Database connected", slog.String("database", cfg.PgDb))

	// // Migrations
	// if err := db.SyncDb(dbConn); err != nil {
	// 	log.Fatalf("DB migration error: %v", err)
	// }

	// Создание репозиториев
	userRepo := repository.NewUserRepository(dbConn)
	authCodeRepo := repository.NewAuthCodeRepository(dbConn)
	orgRepo := repository.NewOrganizationRepository(dbConn)
	locationRepo := repository.NewLocationRepository(dbConn)
	departmentRepo := repository.NewDepartmentRepository(dbConn)
	positionRepo := repository.NewPositionRepository(dbConn)
	permissionRepo := repository.NewPermissionRepository(dbConn)
	permissionGrantRepo := repository.NewPermissionGrantRepository(dbConn)
	employeeRepo := repository.NewEmployeeRepository(dbConn)
	inviteRepo := repository.NewInviteRepository(dbConn)

	// Создание email сервиса
	emailService := services.NewEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.FromEmail,
		cfg.FromName,
	)

	// Создание сервиса аутентификации
	authService := services.NewAuthService(
		userRepo,
		authCodeRepo,
		emailService,
		cfg.JWTSecret,
	)

	// Создание сервиса прав
	permissionService := services.NewPermissionService(permissionRepo, permissionGrantRepo, orgRepo, employeeRepo)

	// Создание сервиса организаций
	orgService := services.NewOrganizationService(orgRepo)

	// Создание сервиса локаций (без permissionSvc)
	locationService := services.NewLocationService(locationRepo, orgRepo)

	// Создание сервиса отделов (без permissionSvc)
	departmentService := services.NewDepartmentService(departmentRepo, locationRepo)

	// Создание сервиса позиций (без permissionSvc)
	positionService := services.NewPositionService(positionRepo)

	// Создание сервиса сотрудников (без permissionSvc)
	employeeService := services.NewEmployeeService(employeeRepo, orgRepo, positionRepo)

	// Создание сервиса приглашений
	inviteService := services.NewInviteService(
		inviteRepo,
		orgRepo,
		positionRepo,
		userRepo,
		employeeRepo,
		permissionService,
		emailService,
		logger.Log,
	)

	// TODO: Create controllers for employeeService
	_ = employeeService

	// Создание middleware
	permMiddleware := middleware.NewPermissionMiddleware(permissionService)

	// Создание контроллеров
	authController := controllers.NewAuthController(authService)
	orgController := controllers.NewOrganizationController(orgService)
	locationController := controllers.NewLocationController(locationService)
	departmentController := controllers.NewDepartmentController(departmentService)
	positionController := controllers.NewPositionController(positionService)
	inviteController := controllers.NewInviteController(inviteService)

	logger.Info("✅ Services initialized", slog.String("port", cfg.AppPort))

	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Добавляем middleware
	// r.Use(middleware.CORS())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.ValidateJSON())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// ===== Публичные роуты аутентификации =====
		auth := api.Group("/auth")
		{
			auth.POST("/request-code", authController.RequestCode)
			auth.POST("/verify-code", authController.VerifyCode)
			auth.POST("/login-password", authController.LoginWithPassword)
		}

		// ===== Роуты требующие JWT (любой статус) =====
		authRequired := api.Group("/")
		authRequired.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			// Выход доступен всегда
			authRequired.POST("/auth/logout", authController.Logout)

			// Заполнение профиля (только для pending_profile + jwt token)
			authRequired.POST("/auth/complete-profile",
				middleware.RequireStatus(models.UserStatusPendingProfile),
				authController.CompleteProfile)
		}

		// ===== Защищённые роуты (только active пользователи) =====
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(cfg.JWTSecret))
		protected.Use(middleware.RequireActiveUser())
		{
			// Профиль
			protected.GET("/profile", authController.GetProfile)
			protected.POST("/profile/set-password", authController.SetPassword)
			protected.POST("/profile/change-password", authController.ChangePassword)
			protected.POST("/profile/remove-password", authController.RemovePassword)

			// Организации
			protected.POST("/organizations", orgController.CreateOrganization)
			protected.GET("/organizations", orgController.GetUserOrganizations)
			protected.GET("/organizations/:orgId",
				permMiddleware.RequireOrgAccess,
				orgController.GetOrganization)
			protected.PUT("/organizations/:orgId",
				permMiddleware.RequireOrgAccess,
				orgController.UpdateOrganization)
			protected.DELETE("/organizations/:orgId",
				permMiddleware.RequireOrgAccess,
				orgController.DeleteOrganization)

			// Локации (with hierarchical permission checking)
			protected.POST("/organizations/:orgId/locations",
				permMiddleware.RequireScopedPermissionHierarchy(
					"locations.create",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				locationController.CreateLocation)
			protected.GET("/organizations/:orgId/locations",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				locationController.GetOrganizationLocations)
			protected.GET("/organizations/:orgId/locations/:locId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeLocation,
					extractLocationContext,
				),
				locationController.GetLocation)
			protected.PUT("/organizations/:orgId/locations/:locId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"locations.update",
					models.ScopeLocation,
					extractLocationContext,
				),
				locationController.UpdateLocation)
			protected.DELETE("/organizations/:orgId/locations/:locId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"locations.delete",
					models.ScopeLocation,
					extractLocationContext,
				),
				locationController.DeleteLocation)

			// Отделы (with hierarchical permission checking)
			protected.POST("/organizations/:orgId/locations/:locId/departments",
				permMiddleware.RequireScopedPermissionHierarchy(
					"departments.create",
					models.ScopeDepartment,
					extractDepartmentCreationContext,
				),
				departmentController.CreateDepartment)
			protected.GET("/organizations/:orgId/locations/:locId/departments",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeLocation,
					extractLocationContext,
				),
				departmentController.GetLocationDepartments)
			protected.GET("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeDepartment,
					extractDepartmentContext,
				),
				departmentController.GetDepartment)
			protected.PUT("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"departments.update",
					models.ScopeDepartment,
					extractDepartmentContext,
				),
				departmentController.UpdateDepartment)
			protected.DELETE("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"departments.delete",
					models.ScopeDepartment,
					extractDepartmentContext,
				),
				departmentController.DeleteDepartment)

			// Позиции (with hierarchical permission checking)
			protected.POST("/organizations/:orgId/positions",
				permMiddleware.RequireScopedPermissionHierarchy(
					"positions.create",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				positionController.CreatePosition)
			protected.GET("/organizations/:orgId/positions",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				positionController.GetOrganizationPositions)
			protected.GET("/organizations/:orgId/locations/:locId/departments/:deptId/positions",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeDepartment,
					extractDepartmentContext,
				),
				positionController.GetDepartmentPositions)
			protected.GET("/organizations/:orgId/positions/:posId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"readHierarchy",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				positionController.GetPosition)
			protected.PUT("/organizations/:orgId/positions/:posId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"positions.update",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				positionController.UpdatePosition)
			protected.DELETE("/organizations/:orgId/positions/:posId",
				permMiddleware.RequireScopedPermissionHierarchy(
					"positions.delete",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				positionController.DeletePosition)

			// Приглашения
			protected.POST("/organizations/:orgId/invites",
				permMiddleware.RequireScopedPermissionHierarchy(
					"invites.create",
					models.ScopeOrganization,
					extractOrganizationContext,
				),
				inviteController.CreateInvite)

			protected.GET("/invites/my",
				inviteController.GetMyInvites)
			protected.POST("/invites/:inviteId/accept",
				inviteController.AcceptInvite)
			protected.POST("/invites/:inviteId/decline",
				inviteController.DeclineInvite)
			protected.DELETE("/invites/:inviteId",
				inviteController.CancelInvite)
			protected.GET("/organizations/:orgId/invites",
				permMiddleware.RequireOrgAccess,
				inviteController.GetOrganizationInvites)
		}
	}

	logger.Info("✅ Server ready", slog.String("url", "http://localhost:"+cfg.AppPort))
	if err := r.Run(":" + cfg.AppPort); err != nil {
		logger.Error("Server failed to start", err)
		panic(err)
	}
}

// Helper functions to extract permission context from URL parameters

// extractOrganizationContext extracts organization-level context
func extractOrganizationContext(c *gin.Context) *models.PermissionContext {
	orgID, _ := strconv.ParseInt(c.Param("orgId"), 10, 64)
	return &models.PermissionContext{
		OrgID: &orgID,
	}
}

// extractLocationContext extracts location-level context
func extractLocationContext(c *gin.Context) *models.PermissionContext {
	permCtx := extractOrganizationContext(c)
	locID, _ := strconv.ParseInt(c.Param("locId"), 10, 64)
	permCtx.LocationID = &locID
	return permCtx
}

// extractDepartmentContext extracts department-level context (includes location)
func extractDepartmentContext(c *gin.Context) *models.PermissionContext {
	permCtx := extractLocationContext(c)
	deptID, _ := strconv.ParseInt(c.Param("deptId"), 10, 64)
	permCtx.DepartmentID = &deptID
	return permCtx
}

func extractDepartmentCreationContext(c *gin.Context) *models.PermissionContext {
	permCtx := extractLocationContext(c)

	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return permCtx
	}

	permCtx.DepartmentID = req.ParentID
	return permCtx
}

package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/config"
	"github.com/osi-oss/osi/internal/controllers"
	"github.com/osi-oss/osi/internal/db"
	"github.com/osi-oss/osi/internal/middleware"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/services"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Start(cfg *config.Config) {
	// Reading config && Connection to db
	dbConn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

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

	// Создание сервиса членов организации (без permissionSvc)
	memberService := services.NewMemberService(orgRepo, userRepo)

	// Создание сервиса сотрудников (без permissionSvc)
	employeeService := services.NewEmployeeService(employeeRepo, orgRepo, positionRepo)

	// TODO: Create controllers for memberService and employeeService
	_ = memberService
	_ = employeeService

	// Создание middleware
	permMiddleware := middleware.NewPermissionMiddleware(permissionService)

	// Создание контроллеров
	authController := controllers.NewAuthController(authService)
	orgController := controllers.NewOrganizationController(orgService)
	locationController := controllers.NewLocationController(locationService)
	departmentController := controllers.NewDepartmentController(departmentService)
	positionController := controllers.NewPositionController(positionService)

	log.Printf("🚀 Server starting on port %s", cfg.AppPort)
	log.Printf("📊 Database: %s@%s:%s/%s", cfg.PgUser, cfg.PgHost, cfg.PgPort, cfg.PgDb)

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
			auth.POST("/resend-code", authController.ResendCode)
			auth.POST("/login-password", authController.LoginWithPassword)
		}

		// ===== Роуты требующие JWT (любой статус) =====
		authRequired := api.Group("/")
		authRequired.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			// Выход доступен всегда
			authRequired.POST("/auth/logout", authController.Logout)

			// Заполнение профиля (только для pending_profile)
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

			// Локации
			protected.POST("/organizations/:orgId/locations",
				permMiddleware.RequirePermission("locations.create"),
				locationController.CreateLocation)
			protected.GET("/organizations/:orgId/locations",
				permMiddleware.RequireOrgAccess,
				locationController.GetOrganizationLocations)
			protected.GET("/organizations/:orgId/locations/:locId",
				permMiddleware.RequireOrgAccess,
				locationController.GetLocation)
			protected.PUT("/organizations/:orgId/locations/:locId",
				permMiddleware.RequirePermission("locations.update"),
				locationController.UpdateLocation)
			protected.DELETE("/organizations/:orgId/locations/:locId",
				permMiddleware.RequirePermission("locations.delete"),
				locationController.DeleteLocation)

			// Отделы
			protected.POST("/organizations/:orgId/locations/:locId/departments",
				permMiddleware.RequirePermission("departments.create"),
				departmentController.CreateDepartment)
			protected.GET("/organizations/:orgId/locations/:locId/departments",
				permMiddleware.RequireOrgAccess,
				departmentController.GetLocationDepartments)
			protected.GET("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequireOrgAccess,
				departmentController.GetDepartment)
			protected.PUT("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequirePermission("departments.update"),
				departmentController.UpdateDepartment)
			protected.DELETE("/organizations/:orgId/locations/:locId/departments/:deptId",
				permMiddleware.RequirePermission("departments.delete"),
				departmentController.DeleteDepartment)

			// Позиции
			protected.POST("/organizations/:orgId/positions",
				permMiddleware.RequirePermission("positions.create"),
				positionController.CreatePosition)
			protected.GET("/organizations/:orgId/positions",
				permMiddleware.RequireOrgAccess,
				positionController.GetOrganizationPositions)
			protected.GET("/organizations/:orgId/locations/:locId/departments/:deptId/positions",
				permMiddleware.RequireOrgAccess,
				positionController.GetDepartmentPositions)
			protected.GET("/organizations/:orgId/positions/:posId",
				permMiddleware.RequireOrgAccess,
				positionController.GetPosition)
			protected.PUT("/organizations/:orgId/positions/:posId",
				permMiddleware.RequirePermission("positions.update"),
				positionController.UpdatePosition)
			protected.DELETE("/organizations/:orgId/positions/:posId",
				permMiddleware.RequirePermission("positions.delete"),
				positionController.DeletePosition)
		}
	}

	log.Printf("✅ Server ready at http://localhost:%s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

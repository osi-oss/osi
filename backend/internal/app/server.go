package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/config"
	"github.com/osi-oss/osi/internal/controllers"
	"github.com/osi-oss/osi/internal/db"
	"github.com/osi-oss/osi/internal/middleware"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/services"
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
	resetRepo := repository.NewPasswordResetRepository(dbConn)
	orgRepo := repository.NewOrganizationRepository(dbConn)
	locationRepo := repository.NewLocationRepository(dbConn)
	departmentRepo := repository.NewDepartmentRepository(dbConn)
	positionRepo := repository.NewPositionRepository(dbConn)
	permissionRepo := repository.NewPermissionRepository(dbConn)
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

	// Создание пользовательского сервиса со всеми зависимостями
	userService := services.NewUserService(
		userRepo,
		resetRepo,
		emailService,
		cfg.JWTSecret,
		cfg.BaseURL,
	)

	// Создание сервиса организаций
	orgService := services.NewOrganizationService(orgRepo)

	// Создание сервиса прав (требует orgRepo и employeeRepo)
	permissionService := services.NewPermissionService(permissionRepo, orgRepo, employeeRepo)

	// Установка PermissionService в OrganizationService (избегаем циклической зависимости)
	orgService.SetPermissionService(permissionService)

	// Создание сервиса локаций
	locationService := services.NewLocationService(locationRepo, orgService, permissionService)

	// Создание сервиса отделов
	departmentService := services.NewDepartmentService(departmentRepo, locationService, permissionService)

	// Создание сервиса позиций
	positionService := services.NewPositionService(positionRepo, orgService, permissionService)

	// Создание сервиса членов организации
	memberService := services.NewMemberService(orgRepo, userRepo, permissionService)

	// Создание сервиса сотрудников
	employeeService := services.NewEmployeeService(employeeRepo, orgRepo, positionRepo, permissionService)

	// TODO: Create controllers for memberService and employeeService
	_ = memberService
	_ = employeeService

	userController := controllers.NewUserController(userService)
	orgController := controllers.NewOrganizationController(orgService)
	locationController := controllers.NewLocationController(locationService)
	departmentController := controllers.NewDepartmentController(departmentService)
	positionController := controllers.NewPositionController(positionService)

	log.Printf("🚀 Server starting on port %s", cfg.AppPort)
	log.Printf("📊 Database: %s@%s:%s/%s", cfg.PgUser, cfg.PgHost, cfg.PgPort, cfg.PgDb)

	r := gin.Default()

	// Добавляем middleware
	// r.Use(middleware.CORS())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.ValidateJSON())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})

	// В server.go добавить:
	api := r.Group("/api")
	{
		// Открытые роуты
		api.POST("/signup", userController.SignUp)
		api.POST("/login", userController.LogIn)
		api.POST("/forgot-password", userController.RequestPasswordReset)

		// Сброс пароля
		api.GET("/reset-password/validate", userController.ValidateResetToken) // Проверка токена
		api.POST("/reset-password", userController.ResetPassword)              // Сброс пароля

		// Защищенные роуты
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			protected.GET("/profile", userController.GetProfile)
			protected.POST("/logout", userController.Logout)

			// Роуты организаций
			protected.POST("/organizations", orgController.CreateOrganization)
			protected.GET("/organizations", orgController.GetUserOrganizations)
			protected.GET("/organizations/:id", orgController.GetOrganization)
			protected.PUT("/organizations/:id", orgController.UpdateOrganization)
			protected.DELETE("/organizations/:id", orgController.DeleteOrganization)

			// Роуты локаций
			protected.POST("/organizations/:id/locations", locationController.CreateLocation)
			protected.GET("/organizations/:id/locations", locationController.GetOrganizationLocations)
			protected.GET("/locations/:id", locationController.GetLocation)
			protected.PUT("/locations/:id", locationController.UpdateLocation)
			protected.DELETE("/locations/:id", locationController.DeleteLocation)

			// Роуты отделов
			protected.POST("/locations/:id/departments", departmentController.CreateDepartment)
			protected.GET("/locations/:id/departments", departmentController.GetLocationDepartments)
			protected.GET("/departments/:id", departmentController.GetDepartment)
			protected.PUT("/departments/:id", departmentController.UpdateDepartment)
			protected.DELETE("/departments/:id", departmentController.DeleteDepartment)

			// Роуты позиций
			protected.POST("/organizations/:id/positions", positionController.CreatePosition)
			protected.GET("/organizations/:id/positions", positionController.GetOrganizationPositions)
			protected.GET("/departments/:id/positions", positionController.GetDepartmentPositions)
			protected.GET("/positions/:id", positionController.GetPosition)
			protected.PUT("/positions/:id", positionController.UpdatePosition)
			protected.DELETE("/positions/:id", positionController.DeletePosition)
		}
	}

	log.Printf("✅ Server ready at http://localhost:%s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

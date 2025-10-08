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

	// Migrations
	if err := db.SyncDb(dbConn); err != nil {
		log.Fatalf("DB migration error: %v", err)
	}

	// Создание репозиториев
	userRepo := repository.NewUserRepository(dbConn)
	resetRepo := repository.NewPasswordResetRepository(dbConn)

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

	userController := controllers.NewUserController(userService)

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
		}
	}

	log.Printf("✅ Server ready at http://localhost:%s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

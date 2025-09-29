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

	userRepo := repository.NewUserRepository(dbConn)
	userService := services.NewUserService(userRepo, cfg.JWTSecret)
	userController := controllers.NewUserController(userService)

	log.Printf("🚀 Server starting on port %s", cfg.AppPort)
	log.Printf("📊 Database: %s@%s:%s/%s", cfg.PgUser, cfg.PgHost, cfg.PgPort, cfg.PgDb)

	r := gin.Default()

	// Добавляем middleware
	r.Use(middleware.CORS())
	r.Use(middleware.ValidateJSON())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api")
	{
		// Открытые роуты (без аутентификации)
		api.POST("/signup", userController.SignUp)
		api.POST("/login", userController.LogIn)

		// Защищенные роуты (с аутентификацией)
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

package main

import (
	server "github.com/osi-oss/osi/internal/app"
	"github.com/osi-oss/osi/internal/config"

	_ "github.com/osi-oss/osi/docs" // swagger docs
)

// @title           OSI API
// @version         1.0
// @description     API для управления организациями, локациями, отделами и позициями
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.LoadFromEnv()

	server.Start(&cfg)
}

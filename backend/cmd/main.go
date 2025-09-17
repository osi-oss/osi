package main

import (
	server "github.com/osi-oss/osi/internal/app"
	"github.com/osi-oss/osi/internal/config"
)

func main() {
	cfg := config.LoadFromEnv()

	server.Start(&cfg)
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
}

func Load(paths ...string) Config {
	if len(paths) == 0 {
		paths = []string{"../.env"}
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err != nil {
			log.Printf("Не удалось загрузить %s: %v", path, err)
		}
	}

	cfg := Config{}
	cfg.AppPort = getEnv("BACKEND_PORT", "8080")
	return cfg
}

func getEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	return v
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env %s", key)
	}
	return v
}

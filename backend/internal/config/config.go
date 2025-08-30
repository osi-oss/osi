package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	PgHost     string
	PgDb       string
	PgUser     string
	PgPassword string
	PgPort     string
}

func Load(paths ...string) Config {
	if len(paths) == 0 {
		// Сначала пытаемся загрузить из папки backend, потом из корня
		paths = []string{".env", "../.env"}
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err != nil {
			log.Printf("Не удалось загрузить %s: %v", path, err)
		}
	}

	cfg := Config{}
	cfg.AppPort = getEnv("BACKEND_PORT", "8080")
	cfg.PgHost = mustEnv("POSTGRES_HOST")
	cfg.PgDb = mustEnv("POSTGRES_DB")
	cfg.PgUser = mustEnv("POSTGRES_USER")
	cfg.PgPassword = mustEnv("POSTGRES_PASSWORD")
	cfg.PgPort = mustEnv("POSTGRES_PORT")
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

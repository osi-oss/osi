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
	JWTSecret  string

	// Email настройки
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string // URL фронтенда для ссылок в письмах
}

func LoadFromEnv(paths ...string) Config {
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
	cfg.JWTSecret = mustEnv("JWT_SECRET")

	// Email настройки
	cfg.SMTPHost = getEnv("SMTP_HOST", "smtp.gmail.com")
	cfg.SMTPPort = getEnv("SMTP_PORT", "587")
	cfg.SMTPUser = getEnv("SMTP_USER", "")
	cfg.SMTPPassword = getEnv("SMTP_PASSWORD", "")
	cfg.FromEmail = getEnv("FROM_EMAIL", cfg.SMTPUser)
	cfg.FromName = getEnv("FROM_NAME", "OSI Team")
	cfg.BaseURL = getEnv("BASE_URL", "http://localhost:3000")

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

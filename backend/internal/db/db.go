package db

import (
	"fmt"

	"github.com/osi-oss/osi/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Format connection string
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.PgUser, cfg.PgPassword, cfg.PgDb, cfg.PgHost, cfg.PgPort)
	return openConnection(dsn)
}

// Format connection string
func ConnectWithParams(user, password, dbname, host, port string) (*gorm.DB, error) {
	// dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
	// 	user, password, dbname, host, port)
	// return openConnection(dsn)

	return Connect(
		&config.Config{
			PgUser:     user,
			PgPassword: password,
			PgHost:     host,
			PgPort:     port,
			PgDb:       dbname,
		})
}

// Open connection with db
func openConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return db, nil
}

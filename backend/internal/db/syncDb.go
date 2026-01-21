package db

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

func SyncDb(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationFounder{},
		&models.OrganizationMember{},
		&models.Location{},
		&models.Department{},
		&models.Position{},
		&models.Employee{},
		&models.Permission{},
		&models.AuthCode{},
	)
}

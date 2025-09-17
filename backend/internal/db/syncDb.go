package db

import (
	"log"

	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

func SyncDb(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{}, &models.Organisation{})

	if err != nil {
		log.Print("failed to connect to db:", err)
	}

}

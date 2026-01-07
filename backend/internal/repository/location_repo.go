package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type LocationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) Create(location *models.Location) error {
	return r.db.Create(location).Error
}

func (r *LocationRepository) GetByID(id int64) (*models.Location, error) {
	var location models.Location
	err := r.db.Preload("Departments").First(&location, id).Error
	return &location, err
}

func (r *LocationRepository) GetByOrganizationID(orgID int64) ([]models.Location, error) {
	var locations []models.Location
	err := r.db.Where("organization_id = ?", orgID).Preload("Departments").Find(&locations).Error
	return locations, err
}

func (r *LocationRepository) Update(location *models.Location) error {
	return r.db.Save(location).Error
}

func (r *LocationRepository) Delete(id int64) error {
	return r.db.Delete(&models.Location{}, id).Error
}

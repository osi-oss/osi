package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(department *models.Department) error {
	return r.db.Create(department).Error
}

func (r *DepartmentRepository) GetByID(id int64) (*models.Department, error) {
	var department models.Department
	err := r.db.Preload("Location").Preload("Parent").Preload("Children").Preload("Positions").First(&department, id).Error
	return &department, err
}

func (r *DepartmentRepository) GetByLocationID(locationID int64) ([]models.Department, error) {
	var departments []models.Department
	err := r.db.Where("location_id = ?", locationID).Preload("Parent").Preload("Children").Find(&departments).Error
	return departments, err
}

func (r *DepartmentRepository) Update(department *models.Department) error {
	return r.db.Save(department).Error
}

func (r *DepartmentRepository) Delete(id int64) error {
	return r.db.Delete(&models.Department{}, id).Error
}

package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type PositionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(db *gorm.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

func (r *PositionRepository) Create(position *models.Position) error {
	return r.db.Create(position).Error
}

func (r *PositionRepository) GetByID(id int64) (*models.Position, error) {
	var position models.Position
	err := r.db.First(&position, id).Error
	return &position, err
}

func (r *PositionRepository) Update(position *models.Position) error {
	return r.db.Save(position).Error
}

func (r *PositionRepository) Delete(id int64) error {
	return r.db.Delete(&models.Position{}, id).Error
}

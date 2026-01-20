package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) GetById(id int64) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// Update обновляет пользователя
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

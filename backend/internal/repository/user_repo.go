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

// UpdateStatus обновляет статус пользователя
func (r *UserRepository) UpdateStatus(userID int64, status models.UserStatus) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

// UpdateProfile обновляет профиль пользователя (имя, фамилия, отчество)
func (r *UserRepository) UpdateProfile(userID int64, firstName, lastName string, middleName *string) error {
	updates := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
	}
	if middleName != nil {
		updates["middle_name"] = *middleName
	}
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}

// SetPassword устанавливает хеш пароля
func (r *UserRepository) SetPassword(userID int64, passwordHash string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// VerifyEmail устанавливает флаг подтверждения email
func (r *UserRepository) VerifyEmail(userID int64) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_email_verified", true).Error
}

// ExistsByEmail проверяет существование пользователя по email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

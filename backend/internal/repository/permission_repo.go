package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *PermissionRepository) GetByID(id int64) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.First(&permission, id).Error
	return &permission, err
}

func (r *PermissionRepository) GetByCode(code string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("code = ?", code).First(&permission).Error
	return &permission, err
}

func (r *PermissionRepository) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) GetByGroupName(groupName string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Where("group_name = ?", groupName).Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

func (r *PermissionRepository) Delete(id int64) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

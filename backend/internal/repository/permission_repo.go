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

// Position permissions methods

func (r *PermissionRepository) AssignPermissionToPosition(positionID int64, permissionID int64) error {
	return r.db.Exec("INSERT INTO position_permissions (position_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING", positionID, permissionID).Error
}

func (r *PermissionRepository) RemovePermissionFromPosition(positionID int64, permissionID int64) error {
	return r.db.Exec("DELETE FROM position_permissions WHERE position_id = ? AND permission_id = ?", positionID, permissionID).Error
}

func (r *PermissionRepository) GetPositionPermissions(positionID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN position_permissions ON permissions.id = position_permissions.permission_id").
		Where("position_permissions.position_id = ?", positionID).
		Find(&permissions).Error
	return permissions, err
}

// Member permissions methods

func (r *PermissionRepository) AssignPermissionToMember(memberID int64, permissionID int64) error {
	return r.db.Exec("INSERT INTO member_permissions (member_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING", memberID, permissionID).Error
}

func (r *PermissionRepository) RemovePermissionFromMember(memberID int64, permissionID int64) error {
	return r.db.Exec("DELETE FROM member_permissions WHERE member_id = ? AND permission_id = ?", memberID, permissionID).Error
}

func (r *PermissionRepository) GetMemberPermissions(memberID int64) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN member_permissions ON permissions.id = member_permissions.permission_id").
		Where("member_permissions.member_id = ?", memberID).
		Find(&permissions).Error
	return permissions, err
}

// Bulk operations

func (r *PermissionRepository) SetPositionPermissions(positionID int64, permissionIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove all existing permissions
		if err := tx.Exec("DELETE FROM position_permissions WHERE position_id = ?", positionID).Error; err != nil {
			return err
		}

		// Add new permissions
		if len(permissionIDs) > 0 {
			for _, permID := range permissionIDs {
				if err := tx.Exec("INSERT INTO position_permissions (position_id, permission_id) VALUES (?, ?)", positionID, permID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *PermissionRepository) SetMemberPermissions(memberID int64, permissionIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove all existing permissions
		if err := tx.Exec("DELETE FROM member_permissions WHERE member_id = ?", memberID).Error; err != nil {
			return err
		}

		// Add new permissions
		if len(permissionIDs) > 0 {
			for _, permID := range permissionIDs {
				if err := tx.Exec("INSERT INTO member_permissions (member_id, permission_id) VALUES (?, ?)", memberID, permID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

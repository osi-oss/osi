package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type InviteRepository struct {
	db *gorm.DB
}

func NewInviteRepository(db *gorm.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

// Create создаёт новое приглашение
func (r *InviteRepository) Create(invite *models.Invite) error {
	return r.db.Create(invite).Error
}

// GetByID получает приглашение по ID
func (r *InviteRepository) GetByID(id int64) (*models.Invite, error) {
	var invite models.Invite
	err := r.db.
		Preload("Organization").
		Preload("Position").
		Preload("InvitedUser").
		Preload("InvitedByUser").
		First(&invite, id).Error
	return &invite, err
}

// GetByEmail получает все приглашения для email
func (r *InviteRepository) GetByEmail(email string) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.
		Where("invited_email = ?", email).
		Preload("Organization").
		Preload("Position").
		Preload("InvitedByUser").
		Find(&invites).Error
	return invites, err
}

// GetByUserID получает все приглашения для пользователя
func (r *InviteRepository) GetByUserID(userID int64) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.
		Where("invited_user_id = ?", userID).
		Preload("Organization").
		Preload("Position").
		Preload("InvitedByUser").
		Find(&invites).Error
	return invites, err
}

// GetPendingByOrgAndPosition получает pending приглашение на должность
func (r *InviteRepository) GetPendingByOrgAndPosition(orgID, positionID, userID int64) (*models.Invite, error) {
	var invite models.Invite
	err := r.db.
		Where("organization_id = ? AND position_id = ? AND invited_user_id = ? AND status = ?",
			orgID, positionID, userID, models.InvitePending).
		First(&invite).Error
	return &invite, err
}

// GetPendingByEmail получает все pending приглашения по email
func (r *InviteRepository) GetPendingByEmail(email string) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.
		Where("invited_email = ? AND status = ?", email, models.InvitePending).
		Preload("Organization").
		Preload("Position").
		Preload("InvitedByUser").
		Find(&invites).Error
	return invites, err
}

// GetByOrganization получает все приглашения организации
func (r *InviteRepository) GetByOrganization(orgID int64) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.
		Where("organization_id = ?", orgID).
		Preload("InvitedUser").
		Preload("Position").
		Preload("InvitedByUser").
		Find(&invites).Error
	return invites, err
}

// GetByOrganizationAndStatus получает приглашения по организации и статусу
func (r *InviteRepository) GetByOrganizationAndStatus(orgID int64, status models.InviteStatus) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.
		Where("organization_id = ? AND status = ?", orgID, status).
		Preload("InvitedUser").
		Preload("Position").
		Preload("InvitedByUser").
		Find(&invites).Error
	return invites, err
}

// UpdateStatus обновляет статус приглашения
func (r *InviteRepository) UpdateStatus(inviteID int64, status models.InviteStatus) error {
	result := r.db.Model(&models.Invite{}).
		Where("id = ?", inviteID).
		Update("status", status)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// UpdateStatusWithTimestamp обновляет статус и соответствующий timestamp
func (r *InviteRepository) UpdateStatusWithTimestamp(inviteID int64, status models.InviteStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// Устанавливаем соответствующий timestamp
	if status == models.InviteAccepted {
		updates["accepted_at"] = gorm.Expr("now()")
	} else if status == models.InviteDeclined {
		updates["declined_at"] = gorm.Expr("now()")
	}

	result := r.db.Model(&models.Invite{}).
		Where("id = ?", inviteID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// Delete удаляет приглашение
func (r *InviteRepository) Delete(inviteID int64) error {
	return r.db.Delete(&models.Invite{}, inviteID).Error
}

// ExistsPendingInvite проверяет наличие pending приглашения по email
func (r *InviteRepository) ExistsPendingInviteByEmail(orgID, positionID int64, email string) (bool, error) {
	var count int64
	err := r.db.
		Where("organization_id = ? AND position_id = ? AND invited_email = ? AND status = ?",
			orgID, positionID, email, models.InvitePending).
		Model(&models.Invite{}).
		Count(&count).Error
	return count > 0, err
}

// ExistsPendingInvite проверяет наличие pending приглашения по userID
func (r *InviteRepository) ExistsPendingInvite(orgID, positionID int64, userID *int64) (bool, error) {
	if userID == nil {
		return false, nil
	}
	var count int64
	err := r.db.
		Where("organization_id = ? AND position_id = ? AND invited_user_id = ? AND status = ?",
			orgID, positionID, *userID, models.InvitePending).
		Model(&models.Invite{}).
		Count(&count).Error
	return count > 0, err
}

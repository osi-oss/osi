package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

func (r *OrganizationRepository) GetByID(id int64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Founders").Preload("Founders.User").Preload("Members").First(&org, id).Error
	return &org, err
}

func (r *OrganizationRepository) GetByUserID(userID int64) ([]models.Organization, error) {
	var organizations []models.Organization
	err := r.db.
		Joins("JOIN organization_founders ON organizations.id = organization_founders.organization_id").
		Where("organization_founders.user_id = ?", userID).
		Preload("Founders").
		Preload("Members").
		Find(&organizations).Error
	return organizations, err
}

func (r *OrganizationRepository) Update(org *models.Organization) error {
	return r.db.Save(org).Error
}

func (r *OrganizationRepository) Delete(id int64) error {
	return r.db.Delete(&models.Organization{}, id).Error
}

func (r *OrganizationRepository) CreateFounder(founder *models.OrganizationFounder) error {
	return r.db.Create(founder).Error
}

func (r *OrganizationRepository) GetFoundersByOrganizationID(orgID int64) ([]models.OrganizationFounder, error) {
	var founders []models.OrganizationFounder
	err := r.db.Where("organization_id = ?", orgID).Preload("User").Find(&founders).Error
	return founders, err
}

// Member methods

func (r *OrganizationRepository) CreateMember(member *models.OrganizationMember) error {
	return r.db.Create(member).Error
}

func (r *OrganizationRepository) GetMembersByOrganizationID(orgID int64) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := r.db.Where("organization_id = ?", orgID).Preload("User").Find(&members).Error
	return members, err
}

func (r *OrganizationRepository) GetMemberByID(memberID int64) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.Preload("User").First(&member, memberID).Error
	return &member, err
}

func (r *OrganizationRepository) UpdateMemberStatus(memberID int64, status models.MemberStatus) error {
	return r.db.Model(&models.OrganizationMember{}).Where("id = ?", memberID).Update("status", status).Error
}

func (r *OrganizationRepository) DeleteMember(memberID int64) error {
	return r.db.Delete(&models.OrganizationMember{}, memberID).Error
}

// GetMemberByUserAndOrgID returns a member by user_id and organization_id
func (r *OrganizationRepository) GetMemberByUserAndOrgID(userID int64, orgID int64) (*models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
		Preload("User").
		Preload("Permissions").
		Preload("Employees").
		Preload("Employees.Position").
		Preload("Employees.Position.Permissions").
		First(&member).Error
	return &member, err
}

// GetFounderByUserAndOrgID returns a founder by user_id and organization_id
func (r *OrganizationRepository) GetFounderByUserAndOrgID(userID int64, orgID int64) (*models.OrganizationFounder, error) {
	var founder models.OrganizationFounder
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
		Preload("User").
		First(&founder).Error
	return &founder, err
}

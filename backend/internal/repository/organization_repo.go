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
	err := r.db.Preload("Founders").Preload("Founders.User").Preload("Employees").First(&org, id).Error
	return &org, err
}

func (r *OrganizationRepository) GetByUserID(userID int64) ([]models.Organization, error) {
	var organizations []models.Organization
	err := r.db.
		Joins("JOIN organization_founders ON organizations.id = organization_founders.organization_id").
		Where("organization_founders.user_id = ?", userID).
		Preload("Founders").
		Preload("Employees").
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

// Employee methods (formerly Member methods)

// GetEmployeesByOrganizationID returns all employees in an organization
func (r *OrganizationRepository) GetEmployeesByOrganizationID(orgID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("organization_id = ?", orgID).Preload("User").Preload("Position").Find(&employees).Error
	return employees, err
}

// GetEmployeeByID returns an employee by ID
func (r *OrganizationRepository) GetEmployeeByID(employeeID int64) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("User").Preload("Position").First(&employee, employeeID).Error
	return &employee, err
}

// UpdateEmployeeStatus updates an employee's status
func (r *OrganizationRepository) UpdateEmployeeStatus(employeeID int64, status models.MemberStatus) error {
	return r.db.Model(&models.Employee{}).Where("id = ?", employeeID).Update("status", status).Error
}

// DeleteEmployee deletes an employee
func (r *OrganizationRepository) DeleteEmployee(employeeID int64) error {
	return r.db.Delete(&models.Employee{}, employeeID).Error
}

// GetEmployeeByUserAndOrgID returns an employee of a user in an organization
func (r *OrganizationRepository) GetEmployeeByUserAndOrgID(userID int64, orgID int64) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
		Preload("User").
		Preload("Position").
		First(&employee).Error
	return &employee, err
}

// GetEmployeesByUserAndOrgID returns all employees of a user in an organization
func (r *OrganizationRepository) GetEmployeesByUserAndOrgID(userID int64, orgID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
		Preload("User").
		Preload("Position").
		Find(&employees).Error
	return employees, err
}

// GetFounderByUserAndOrgID returns a founder by user_id and organization_id
func (r *OrganizationRepository) GetFounderByUserAndOrgID(userID int64, orgID int64) (*models.OrganizationFounder, error) {
	var founder models.OrganizationFounder
	err := r.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
		Preload("User").
		First(&founder).Error
	return &founder, err
}

package repository

import (
	"github.com/osi-oss/osi/internal/models"
	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(employee *models.Employee) error {
	return r.db.Create(employee).Error
}

func (r *EmployeeRepository) GetByID(id int64) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.Preload("Member").Preload("Member.User").Preload("Position").First(&employee, id).Error
	return &employee, err
}

func (r *EmployeeRepository) Update(employee *models.Employee) error {
	return r.db.Save(employee).Error
}

func (r *EmployeeRepository) Delete(id int64) error {
	return r.db.Delete(&models.Employee{}, id).Error
}

// GetByUserAndOrgID получает сотрудника по UserID и OrganizationID
func (r *EmployeeRepository) GetByUserAndOrgID(userID, orgID int64) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.
		Preload("User").
		Preload("Position").
		Preload("Organization").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		First(&employee).Error
	return &employee, err
}

// GetByOrganizationID получает всех сотрудников организации
func (r *EmployeeRepository) GetByOrganizationID(orgID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.
		Preload("User").
		Preload("Position").
		Where("organization_id = ?", orgID).
		Find(&employees).Error
	return employees, err
}

// GetByOrganizationIDAndStatus получает сотрудников организации с определённым статусом
func (r *EmployeeRepository) GetByOrganizationIDAndStatus(orgID int64, status models.MemberStatus) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.
		Preload("User").
		Preload("Position").
		Where("organization_id = ? AND status = ?", orgID, status).
		Find(&employees).Error
	return employees, err
}

// GetByOrganizationIDWithPositionInfo получает сотрудников с информацией о позициях
func (r *EmployeeRepository) GetByOrganizationIDWithPositionInfo(orgID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.
		Preload("User").
		Preload("Position", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Department", func(db *gorm.DB) *gorm.DB {
				return db.Preload("Location")
			})
		}).
		Where("organization_id = ?", orgID).
		Order("position_id ASC").
		Find(&employees).Error
	return employees, err
}

// CountByOrganizationID подсчитывает количество сотрудников в организации
func (r *EmployeeRepository) CountByOrganizationID(orgID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Employee{}).Where("organization_id = ?", orgID).Count(&count).Error
	return count, err
}

// CountActiveByOrganizationID подсчитывает активных сотрудников
func (r *EmployeeRepository) CountActiveByOrganizationID(orgID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Employee{}).
		Where("organization_id = ? AND status = ?", orgID, models.MemberActive).
		Where("end_date IS NULL").
		Count(&count).Error
	return count, err
}

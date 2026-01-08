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

func (r *EmployeeRepository) GetByMemberID(memberID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("member_id = ?", memberID).Preload("Position").Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) GetByPositionID(positionID int64) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("position_id = ?", positionID).Preload("Member").Preload("Member.User").Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) Update(employee *models.Employee) error {
	return r.db.Save(employee).Error
}

func (r *EmployeeRepository) Delete(id int64) error {
	return r.db.Delete(&models.Employee{}, id).Error
}

package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type EmployeeService struct {
	employeeRepo *repository.EmployeeRepository
	orgRepo      *repository.OrganizationRepository
	positionRepo *repository.PositionRepository
}

func NewEmployeeService(
	employeeRepo *repository.EmployeeRepository,
	orgRepo *repository.OrganizationRepository,
	positionRepo *repository.PositionRepository,
) *EmployeeService {
	return &EmployeeService{
		employeeRepo: employeeRepo,
		orgRepo:      orgRepo,
		positionRepo: positionRepo,
	}
}

func (s *EmployeeService) GetEmployee(employeeID int64) (*models.Employee, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, apperrors.ErrEmployeeNotFound
	}
	return employee, nil
}

func (s *EmployeeService) UpdateEmployee(employeeID int64, input dto.UpdateEmployeeRequest) (*models.Employee, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, apperrors.ErrEmployeeNotFound
	}

	if input.EndDate != nil {
		employee.EndDate = input.EndDate
	}
	if input.IsIntern != nil {
		employee.IsIntern = *input.IsIntern
	}

	if err := s.employeeRepo.Update(employee); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to update employee")
	}

	return s.employeeRepo.GetByID(employee.ID)
}

func (s *EmployeeService) RemoveFromPosition(employeeID int64) error {
	_, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return apperrors.ErrEmployeeNotFound
	}

	return s.employeeRepo.Delete(employeeID)
}

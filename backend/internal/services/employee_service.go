package services

import (
	"errors"
	"time"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrMemberNotActive  = errors.New("member is not active")
)

type EmployeeService struct {
	employeeRepo  *repository.EmployeeRepository
	orgRepo       *repository.OrganizationRepository
	positionRepo  *repository.PositionRepository
	permissionSvc *PermissionService
}

func NewEmployeeService(
	employeeRepo *repository.EmployeeRepository,
	orgRepo *repository.OrganizationRepository,
	positionRepo *repository.PositionRepository,
	permissionSvc *PermissionService,
) *EmployeeService {
	return &EmployeeService{
		employeeRepo:  employeeRepo,
		orgRepo:       orgRepo,
		positionRepo:  positionRepo,
		permissionSvc: permissionSvc,
	}
}

type AssignPositionInput struct {
	MemberID   int64      `json:"member_id" binding:"required"`
	PositionID int64      `json:"position_id" binding:"required"`
	StartDate  *time.Time `json:"start_date"`
	IsIntern   bool       `json:"is_intern"`
}

type UpdateEmployeeInput struct {
	EndDate  *time.Time `json:"end_date"`
	IsIntern *bool      `json:"is_intern"`
}

// AssignPosition assigns a position to a member (requires positions.create permission)
func (s *EmployeeService) AssignPosition(actorUserID int64, input AssignPositionInput) (*models.Employee, error) {
	// Get the member
	member, err := s.orgRepo.GetMemberByID(input.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, err
	}

	// Check if member is active
	if member.Status != models.MemberActive {
		return nil, ErrMemberNotActive
	}

	// Check if actor has permission to assign positions
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "positions.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	// Verify position exists and belongs to same organization
	position, err := s.positionRepo.GetByID(input.PositionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPositionNotFound
		}
		return nil, err
	}

	if position.OrganizationID != member.OrganizationID {
		return nil, errors.New("position does not belong to the same organization")
	}

	// Create employee record
	employee := &models.Employee{
		MemberID:   input.MemberID,
		PositionID: input.PositionID,
		StartDate:  input.StartDate,
		IsIntern:   input.IsIntern,
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, err
	}

	// Load relations
	employee.Member = *member
	employee.Position = *position

	return employee, nil
}

// GetMemberEmployees returns all employment records for a member
func (s *EmployeeService) GetMemberEmployees(actorUserID int64, memberID int64) ([]models.Employee, error) {
	// Get the member
	member, err := s.orgRepo.GetMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return s.employeeRepo.GetByMemberID(memberID)
}

// GetPositionEmployees returns all employees for a position
func (s *EmployeeService) GetPositionEmployees(actorUserID int64, positionID int64) ([]models.Employee, error) {
	// Get the position
	position, err := s.positionRepo.GetByID(positionID)
	if err != nil {
		return nil, err
	}

	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, position.OrganizationID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return s.employeeRepo.GetByPositionID(positionID)
}

// GetEmployeeByID returns an employee by ID
func (s *EmployeeService) GetEmployeeByID(actorUserID int64, employeeID int64) (*models.Employee, error) {
	// Get the employee
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	// Check if actor has permission to view members
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, employee.Member.OrganizationID, "members.view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	return employee, nil
}

// UpdateEmployee updates an employee record (requires positions.create permission)
func (s *EmployeeService) UpdateEmployee(actorUserID int64, employeeID int64, input UpdateEmployeeInput) (*models.Employee, error) {
	// Get the employee
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	// Check if actor has permission to manage positions
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, employee.Member.OrganizationID, "positions.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	// Update fields
	if input.EndDate != nil {
		employee.EndDate = input.EndDate
	}
	if input.IsIntern != nil {
		employee.IsIntern = *input.IsIntern
	}

	if err := s.employeeRepo.Update(employee); err != nil {
		return nil, err
	}

	return employee, nil
}

// RemoveEmployee deletes an employee record (requires positions.create permission)
func (s *EmployeeService) RemoveEmployee(actorUserID int64, employeeID int64) error {
	// Get the employee
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmployeeNotFound
		}
		return err
	}

	// Check if actor has permission to manage positions
	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, employee.Member.OrganizationID, "positions.create")
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrAccessDenied
	}

	return s.employeeRepo.Delete(employeeID)
}

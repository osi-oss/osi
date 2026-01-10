package services

import (
	"errors"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"gorm.io/gorm"
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

func (s *EmployeeService) AssignPosition(actorUserID int64, input dto.AssignPositionRequest) (*models.Employee, error) {
	member, err := s.orgRepo.GetMemberByID(input.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMemberNotFound
		}
		return nil, err
	}

	if member.Status != models.MemberActive {
		return nil, apperrors.ErrMemberNotActive
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "positions.assign")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
	}

	position, err := s.positionRepo.GetByID(input.PositionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrPositionNotFound
		}
		return nil, err
	}

	if position.OrganizationID != member.OrganizationID {
		return nil, apperrors.BadRequest("position does not belong to the same organization")
	}

	employee := &models.Employee{
		MemberID:   input.MemberID,
		PositionID: input.PositionID,
		StartDate:  input.StartDate,
		IsIntern:   input.IsIntern,
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to assign position")
	}

	return s.employeeRepo.GetByID(employee.ID)
}

func (s *EmployeeService) GetEmployee(employeeID int64, userID int64) (*models.Employee, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, apperrors.ErrEmployeeNotFound
	}

	member, err := s.orgRepo.GetMemberByID(employee.MemberID)
	if err != nil {
		return nil, err
	}

	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, member.OrganizationID)
	if err == nil && founder != nil && founder.ID > 0 {
		return employee, nil
	}

	memberCheck, err := s.orgRepo.GetMemberByUserAndOrgID(userID, member.OrganizationID)
	if err == nil && memberCheck != nil && memberCheck.ID > 0 && memberCheck.Status == models.MemberActive {
		return employee, nil
	}

	return nil, apperrors.ErrForbidden
}

func (s *EmployeeService) UpdateEmployee(actorUserID int64, employeeID int64, input dto.UpdateEmployeeRequest) (*models.Employee, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, apperrors.ErrEmployeeNotFound
	}

	member, err := s.orgRepo.GetMemberByID(employee.MemberID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "positions.assign")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
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

func (s *EmployeeService) RemoveFromPosition(actorUserID int64, employeeID int64) error {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return apperrors.ErrEmployeeNotFound
	}

	member, err := s.orgRepo.GetMemberByID(employee.MemberID)
	if err != nil {
		return err
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(actorUserID, member.OrganizationID, "positions.assign")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	return s.employeeRepo.Delete(employeeID)
}

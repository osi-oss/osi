package services

import (
	"time"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
}

func NewOrganizationService(orgRepo *repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

func (s *OrganizationService) CreateOrganization(userID int64, input dto.CreateOrganizationRequest) (*models.Organization, error) {
	if input.Name == "" {
		return nil, apperrors.BadRequest("organization name is required")
	}

	org := &models.Organization{
		Name:         input.Name,
		LegalName:    input.LegalName,
		INN:          input.INN,
		OGRN:         input.OGRN,
		KPP:          input.KPP,
		LegalAddress: input.LegalAddress,
		Status:       models.OrgDraft,
	}

	if err := s.orgRepo.Create(org); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create organization")
	}

	founder := &models.OrganizationFounder{
		OrganizationID: org.ID,
		UserID:         userID,
		SharePercent:   input.SharePercent,
		IsMain:         true,
	}

	if err := s.orgRepo.CreateFounder(founder); err != nil {
		s.orgRepo.Delete(org.ID)
		return nil, apperrors.Wrap(err, 500, "failed to create founder")
	}

	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) GetOrganization(orgID int64, userID int64) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, apperrors.ErrOrganizationNotFound
	}

	hasAccess := false
	for _, founder := range org.Founders {
		if founder.UserID == userID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		// Check employees (members)
		for _, employee := range org.Employees {
			if employee.UserID == userID && employee.Status == models.MemberActive {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return org, nil
}

func (s *OrganizationService) GetUserOrganizations(userID int64) ([]models.Organization, error) {
	return s.orgRepo.GetByUserID(userID)
}

// UserHasAccessToOrganization checks if user has access to organization
func (s *OrganizationService) UserHasAccessToOrganization(userID int64, orgID int64) (bool, error) {
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		return true, nil
	}

	employee, err := s.orgRepo.GetEmployeeByUserAndOrgID(userID, orgID)
	if err == nil && employee != nil && employee.ID > 0 && employee.Status == models.MemberActive {
		return true, nil
	}

	return false, nil
}

func (s *OrganizationService) UpdateOrganization(orgID int64, userID int64, input dto.UpdateOrganizationRequest) (*models.Organization, error) {
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		org.Name = input.Name
	}
	org.LegalName = input.LegalName
	org.INN = input.INN
	org.OGRN = input.OGRN
	org.KPP = input.KPP
	org.LegalAddress = input.LegalAddress

	if err := s.orgRepo.Update(org); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to update organization")
	}

	return s.orgRepo.GetByID(org.ID)
}

func (s *OrganizationService) DeleteOrganization(orgID int64, userID int64) error {
	org, err := s.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	isMainFounder := false
	for _, founder := range org.Founders {
		if founder.UserID == userID && founder.IsMain {
			isMainFounder = true
			break
		}
	}

	if !isMainFounder {
		return apperrors.ErrForbidden
	}

	return s.orgRepo.Delete(org.ID)
}

// GetUserOrganizationsWithDetails возвращает организации пользователя с деталями
func (s *OrganizationService) GetUserOrganizationsWithDetails(userID int64) ([]dto.MyOrganizationInfo, error) {
	// Получаем организации где пользователь - основатель
	founderOrgs, err := s.orgRepo.GetFounderOrganizations(userID)
	if err != nil && err.Error() != "record not found" {
		return nil, apperrors.Wrap(err, 500, "failed to get founder organizations")
	}

	// Получаем организации где пользователь - сотрудник
	employeeOrgs, err := s.orgRepo.GetOrganizationsByUserID(userID)
	if err != nil && err.Error() != "record not found" {
		return nil, apperrors.Wrap(err, 500, "failed to get employee organizations")
	}

	result := make([]dto.MyOrganizationInfo, 0)

	// Обрабатываем founder organizations
	founderOrgMap := make(map[int64]bool)
	for _, org := range founderOrgs {
		// Попробуем найти employee record
		emp, _ := s.orgRepo.GetEmployeeByUserAndOrgID(userID, org.ID)
		status := "active"
		if emp != nil && emp.ID > 0 {
			status = string(emp.Status)
		}

		result = append(result, dto.MyOrganizationInfo{
			ID:       org.ID,
			Name:     org.Name,
			Status:   string(org.Status),
			MyStatus: status,
			EmployeeID: func() int64 {
				if emp != nil {
					return emp.ID
				}
				return 0
			}(),
			StartDate: func() *time.Time {
				if emp != nil {
					return emp.StartDate
				}
				return nil
			}(),
			EndDate: func() *time.Time {
				if emp != nil {
					return emp.EndDate
				}
				return nil
			}(),
			IsFounder: true,
		})
		founderOrgMap[org.ID] = true
	}

	// Обрабатываем employee organizations (избегаем дубликатов)
	for _, emp := range employeeOrgs {
		if !founderOrgMap[emp.OrganizationID] {
			result = append(result, dto.MyOrganizationInfo{
				ID:         emp.Organization.ID,
				Name:       emp.Organization.Name,
				Status:     string(emp.Organization.Status),
				MyStatus:   string(emp.Status),
				EmployeeID: emp.ID,
				StartDate:  emp.StartDate,
				EndDate:    emp.EndDate,
				IsFounder:  false,
			})
		}
	}

	return result, nil
}

// GetOrganizationEmployees возвращает всех сотрудников организации
func (s *OrganizationService) GetOrganizationEmployees(orgID int64) ([]dto.EmployeeDetailResponse, error) {
	employees, err := s.orgRepo.GetEmployeesByOrganization(orgID)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to get employees")
	}

	result := make([]dto.EmployeeDetailResponse, len(employees))
	for i, emp := range employees {
		result[i] = *dto.ToEmployeeDetailResponse(&emp)
	}
	return result, nil
}

// GetOrganizationHierarchy возвращает всю иерархию организации
func (s *OrganizationService) GetOrganizationHierarchy(orgID int64) (*dto.OrganizationHierarchyResponse, error) {
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, apperrors.ErrOrganizationNotFound
	}

	// Получаем локации
	locations, err := s.orgRepo.GetLocationsByOrganization(orgID)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to get locations")
	}

	// Строим иерархию
	locationNodes := make([]dto.HierarchyNode, len(locations))
	for i, loc := range locations {
		locationNodes[i] = s.buildLocationHierarchy(loc)
	}

	// Считаем статистику
	totalEmployees, _ := s.orgRepo.CountEmployeesInOrganization(orgID)
	totalLocations := int64(len(locations))
	totalDepts, _ := s.orgRepo.CountDepartmentsInOrganization(orgID)
	totalPositions, _ := s.orgRepo.CountPositionsInOrganization(orgID)
	activeEmployees, _ := s.orgRepo.CountActiveEmployeesInOrganization(orgID)

	return &dto.OrganizationHierarchyResponse{
		OrganizationID:   org.ID,
		OrganizationName: org.Name,
		Locations:        locationNodes,
		Stats: struct {
			TotalEmployees   int `json:"total_employees" example:"42"`
			TotalLocations   int `json:"total_locations" example:"3"`
			TotalDepartments int `json:"total_departments" example:"12"`
			TotalPositions   int `json:"total_positions" example:"87"`
			ActiveEmployees  int `json:"active_employees" example:"40"`
		}{
			TotalEmployees:   int(totalEmployees),
			TotalLocations:   int(totalLocations),
			TotalDepartments: int(totalDepts),
			TotalPositions:   int(totalPositions),
			ActiveEmployees:  int(activeEmployees),
		},
	}, nil
}

func (s *OrganizationService) buildLocationHierarchy(loc models.Location) dto.HierarchyNode {
	// Здесь нужно получить отделы для этой локации
	// TODO: реализовать получение отделов через сервис
	return dto.HierarchyNode{
		Type: "location",
		ID:   loc.ID,
		Name: loc.Name,
		Data: map[string]interface{}{
			"source":    loc.Source,
			"is_active": loc.IsActive,
		},
		Children: []dto.HierarchyNode{},
	}
}

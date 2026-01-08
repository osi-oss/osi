package services

import (
	"errors"
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

var (
	ErrLocationNotFound = errors.New("location not found")
)

type LocationService struct {
	locationRepo  *repository.LocationRepository
	orgService    *OrganizationService
	permissionSvc *PermissionService
}

func NewLocationService(locationRepo *repository.LocationRepository, orgService *OrganizationService, permissionSvc *PermissionService) *LocationService {
	return &LocationService{
		locationRepo:  locationRepo,
		orgService:    orgService,
		permissionSvc: permissionSvc,
	}
}

type CreateLocationInput struct {
	Name       string  `json:"name" binding:"required"`
	Address    *string `json:"address"`
	Source     string  `json:"source" binding:"required"` // 'registry' | 'manual'
	IsVerified bool    `json:"is_verified"`
}

type UpdateLocationInput struct {
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	Source     string  `json:"source"`
	IsVerified *bool   `json:"is_verified"`
	IsActive   *bool   `json:"is_active"`
}

func (s *LocationService) CreateLocation(orgID int64, userID int64, input CreateLocationInput) (*models.Location, error) {
	// Check if user has permission to create locations (using departments.create permission)
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, orgID, "departments.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	location := &models.Location{
		OrganizationID: orgID,
		Name:           input.Name,
		Address:        input.Address,
		Source:         input.Source,
		IsVerified:     input.IsVerified,
		IsActive:       true,
	}

	if err := s.locationRepo.Create(location); err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return s.locationRepo.GetByID(location.ID)
}

func (s *LocationService) GetLocation(locationID int64, userID int64) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	// Check if user has access to this organization
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, location.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return location, nil
}

func (s *LocationService) GetOrganizationLocations(orgID int64, userID int64) ([]models.Location, error) {
	// Check if user has access to this organization
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.locationRepo.GetByOrganizationID(orgID)
}

func (s *LocationService) UpdateLocation(locationID int64, userID int64, input UpdateLocationInput) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	// Check if user has permission to update locations
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "departments.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrAccessDenied
	}

	// Обновляем только переданные поля
	if input.Name != "" {
		location.Name = input.Name
	}
	if input.Address != nil {
		location.Address = input.Address
	}
	if input.Source != "" {
		location.Source = input.Source
	}
	if input.IsVerified != nil {
		location.IsVerified = *input.IsVerified
	}
	if input.IsActive != nil {
		location.IsActive = *input.IsActive
	}

	if err := s.locationRepo.Update(location); err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return s.locationRepo.GetByID(location.ID)
}

func (s *LocationService) DeleteLocation(locationID int64, userID int64) error {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return ErrLocationNotFound
	}

	// Check if user has permission to delete locations
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "departments.create")
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrAccessDenied
	}

	return s.locationRepo.Delete(locationID)
}

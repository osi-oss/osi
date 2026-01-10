package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
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

func (s *LocationService) CreateLocation(orgID int64, userID int64, input dto.CreateLocationRequest) (*models.Location, error) {
	hasPermission, err := s.permissionSvc.UserHasPermission(userID, orgID, "locations.create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
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
		return nil, apperrors.Wrap(err, 500, "failed to create location")
	}

	return s.locationRepo.GetByID(location.ID)
}

func (s *LocationService) GetLocation(locationID int64, userID int64) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, apperrors.ErrLocationNotFound
	}

	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, location.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return location, nil
}

func (s *LocationService) GetOrganizationLocations(orgID int64, userID int64) ([]models.Location, error) {
	hasAccess, err := s.orgService.UserHasAccessToOrganization(userID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, apperrors.ErrForbidden
	}

	return s.locationRepo.GetByOrganizationID(orgID)
}

func (s *LocationService) UpdateLocation(locationID int64, userID int64, input dto.UpdateLocationRequest) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, apperrors.ErrLocationNotFound
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "locations.update")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, apperrors.ErrAccessDenied
	}

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
		return nil, apperrors.Wrap(err, 500, "failed to update location")
	}

	return s.locationRepo.GetByID(location.ID)
}

func (s *LocationService) DeleteLocation(locationID int64, userID int64) error {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return apperrors.ErrLocationNotFound
	}

	hasPermission, err := s.permissionSvc.UserHasPermission(userID, location.OrganizationID, "locations.delete")
	if err != nil {
		return err
	}
	if !hasPermission {
		return apperrors.ErrAccessDenied
	}

	return s.locationRepo.Delete(locationID)
}

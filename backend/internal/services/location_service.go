package services

import (
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type LocationService struct {
	locationRepo *repository.LocationRepository
	orgRepo      *repository.OrganizationRepository
}

func NewLocationService(locationRepo *repository.LocationRepository, orgRepo *repository.OrganizationRepository) *LocationService {
	return &LocationService{
		locationRepo: locationRepo,
		orgRepo:      orgRepo,
	}
}

func (s *LocationService) CreateLocation(orgID int64, input dto.CreateLocationRequest) (*models.Location, error) {
	// Check if organization exists
	_, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, apperrors.ErrOrganizationNotFound
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

func (s *LocationService) GetLocation(locationID int64) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, apperrors.ErrLocationNotFound
	}
	return location, nil
}

func (s *LocationService) GetOrganizationLocations(orgID int64) ([]models.Location, error) {
	return s.locationRepo.GetByOrganizationID(orgID)
}

func (s *LocationService) UpdateLocation(locationID int64, input dto.UpdateLocationRequest) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, apperrors.ErrLocationNotFound
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

func (s *LocationService) DeleteLocation(locationID int64) error {
	_, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return apperrors.ErrLocationNotFound
	}

	return s.locationRepo.Delete(locationID)
}

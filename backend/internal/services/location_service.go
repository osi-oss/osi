package services

import (
	"fmt"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
)

type LocationService struct {
	locationRepo *repository.LocationRepository
	orgService   *OrganizationService
}

func NewLocationService(locationRepo *repository.LocationRepository, orgService *OrganizationService) *LocationService {
	return &LocationService{
		locationRepo: locationRepo,
		orgService:   orgService,
	}
}

type CreateLocationInput struct {
	Name       string  `json:"name" binding:"required"`
	Address    *string `json:"address"`
	Source     string  `json:"source" binding:"required"` // 'registry' | 'manual'
	IsVerified bool    `json:"is_verified"`
}

func (s *LocationService) CreateLocation(orgID int64, userID int64, input CreateLocationInput) (*models.Location, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
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

func (s *LocationService) GetLocation(locationID int64, orgID int64, userID int64) (*models.Location, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	return s.locationRepo.GetByID(locationID)
}

func (s *LocationService) GetLocationsByOrganization(orgID int64, userID int64) ([]models.Location, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	return s.locationRepo.GetByOrganizationID(orgID)
}

func (s *LocationService) UpdateLocation(locationID int64, orgID int64, userID int64, input CreateLocationInput) (*models.Location, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	location.Name = input.Name
	location.Address = input.Address
	location.Source = input.Source
	location.IsVerified = input.IsVerified

	if err := s.locationRepo.Update(location); err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return s.locationRepo.GetByID(location.ID)
}

func (s *LocationService) DeleteLocation(locationID int64, orgID int64, userID int64) error {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return err
	}

	return s.locationRepo.Delete(locationID)
}

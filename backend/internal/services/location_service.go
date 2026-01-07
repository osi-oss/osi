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

type UpdateLocationInput struct {
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	Source     string  `json:"source"`
	IsVerified *bool   `json:"is_verified"`
	IsActive   *bool   `json:"is_active"`
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

func (s *LocationService) GetLocation(locationID int64, userID int64) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	// Проверяем доступ к организации этой локации
	_, err = s.orgService.GetOrganization(location.OrganizationID, userID)
	if err != nil {
		return nil, err
	}

	return location, nil
}

func (s *LocationService) GetOrganizationLocations(orgID int64, userID int64) ([]models.Location, error) {
	// Проверяем доступ к организации
	_, err := s.orgService.GetOrganization(orgID, userID)
	if err != nil {
		return nil, err
	}

	return s.locationRepo.GetByOrganizationID(orgID)
}

func (s *LocationService) UpdateLocation(locationID int64, userID int64, input UpdateLocationInput) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(locationID)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	// Проверяем доступ к организации этой локации
	_, err = s.orgService.GetOrganization(location.OrganizationID, userID)
	if err != nil {
		return nil, err
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

	// Проверяем доступ к организации этой локации
	_, err = s.orgService.GetOrganization(location.OrganizationID, userID)
	if err != nil {
		return err
	}

	return s.locationRepo.Delete(locationID)
}

package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const bcryptCost = 12

// Кастомные ошибки
var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input data")
)

type UserService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewUserService(userRepo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{userRepo: userRepo,
		jwtSecret: jwtSecret}
}

func (s *UserService) SignUp(email, password string) (*models.User, error) {
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Verified: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

func (s *UserService) LogIn(email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", fmt.Errorf("user with this email does not exists")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"email": user.Email,
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))

	if err != nil {
		return "", fmt.Errorf("jwt error: %w", err)
	}

	return tokenString, nil
}

// GetByID получает пользователя по ID
func (s *UserService) GetByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetById(id) // исправлено GetById вместо GetByID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}

package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/validators"
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
	userRepo     *repository.UserRepository
	resetRepo    *repository.PasswordResetRepository
	emailService interface {
		SendPasswordResetEmail(toEmail, token, baseURL string) error
		SendWelcomeEmail(toEmail, userName string) error
	}
	jwtSecret string
	baseURL   string
}

func NewUserService(
	userRepo *repository.UserRepository,
	resetRepo *repository.PasswordResetRepository,
	emailService interface {
		SendPasswordResetEmail(toEmail, token, baseURL string) error
		SendWelcomeEmail(toEmail, userName string) error
	},
	jwtSecret string,
	baseURL string,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		resetRepo:    resetRepo,
		emailService: emailService,
		jwtSecret:    jwtSecret,
		baseURL:      baseURL,
	}
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
	// Get user from db
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", fmt.Errorf("user with this email does not exists")
	}

	// Compare password
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

// RequestPasswordReset создает токен и отправляет email для сброса пароля
func (s *UserService) RequestPasswordReset(email string) error {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Не раскрываем что пользователь не найден (безопасность)
			return nil
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Генерируем случайный токен
	token, err := s.generateResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Создаем запись токена в БД
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour), // Токен действует 1 час
		Used:      false,
	}

	if err := s.resetRepo.Create(resetToken); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Отправляем email
	if err := s.emailService.SendPasswordResetEmail(email, token, s.baseURL); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// ResetPassword сбрасывает пароль по токену
func (s *UserService) ResetPassword(token, newPassword string) error {
	// Валидация пароля
	if err := validators.Password.Validate(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Находим токен
	resetToken, err := s.resetRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Получаем пользователя
	user, err := s.userRepo.GetById(resetToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль пользователя
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Помечаем токен как использованный
	if err := s.resetRepo.MarkAsUsed(resetToken.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Удаляем все остальные токены пользователя
	if err := s.resetRepo.DeleteByUserID(user.ID); err != nil {
		// Логируем, но не прерываем процесс
		fmt.Printf("Warning: failed to cleanup reset tokens for user %d: %v\n", user.ID, err)
	}

	return nil
}

// generateResetToken генерирует случайный токен для сброса пароля
func (s *UserService) generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateResetToken проверяет валидность токена без его использования
func (s *UserService) ValidateResetToken(token string) (bool, error) {
	_, err := s.resetRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("database error: %w", err)
	}
	return true, nil
}

package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/validators"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const bcryptCost = 12

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
	// validate password
	if err := validators.Password.Validate(password); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// user already exists
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, apperrors.ErrUserAlreadyExists
	}

	// another error from db
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// validate email
	err = validators.Email.Validate(email)
	if err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to hash password")
	}

	user := &models.User{
		Email:           &email,
		PasswordHash:    string(hashedPassword),
		IsEmailVerified: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create user")
	}

	return user, nil
}

func (s *UserService) LogIn(email, password string) (string, error) {
	// Get user from db
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
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
		return "", apperrors.Wrap(err, 500, "jwt error")
	}

	return tokenString, nil
}

// GetByID получает пользователя по ID
func (s *UserService) GetByID(id int64) (*models.User, error) {
	user, err := s.userRepo.GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, 500, "database error")
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
		return apperrors.Wrap(err, 500, "database error")
	}

	// Генерируем случайный токен
	token, err := s.generateResetToken()
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to generate token")
	}

	// Создаем запись токена в БД
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour), // Токен действует 1 час
		Used:      false,
	}

	if err := s.resetRepo.Create(resetToken); err != nil {
		return apperrors.Wrap(err, 500, "failed to save reset token")
	}

	// Отправляем email
	if err := s.emailService.SendPasswordResetEmail(email, token, s.baseURL); err != nil {
		return apperrors.Wrap(err, 500, "failed to send email")
	}

	return nil
}

// ResetPassword сбрасывает пароль по токену
func (s *UserService) ResetPassword(token, newPassword string) error {
	// Валидация пароля
	if err := validators.Password.Validate(newPassword); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Находим токен
	resetToken, err := s.resetRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BadRequest("invalid or expired reset token")
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetById(resetToken.UserID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to hash password")
	}

	// Обновляем пароль пользователя
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return apperrors.Wrap(err, 500, "failed to update password")
	}

	// Помечаем токен как использованный
	if err := s.resetRepo.MarkAsUsed(resetToken.ID); err != nil {
		return apperrors.Wrap(err, 500, "failed to mark token as used")
	}

	// Удаляем все остальные токены пользователя
	_ = s.resetRepo.DeleteByUserID(user.ID)

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

package services

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/validators"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	bcryptCost           = 12
	jwtExpirationHours   = 24
	maxCodesPerHour      = 5
	codeValidityMinutes  = 20
	codeRateLimitMinutes = 60
)

// AuthService сервис аутентификации
type AuthService struct {
	userRepo     *repository.UserRepository
	authCodeRepo *repository.AuthCodeRepository
	emailService interface {
		SendAuthCode(toEmail, code string) error
	}
	jwtSecret string
}

// NewAuthService создаёт новый сервис аутентификации
func NewAuthService(
	userRepo *repository.UserRepository,
	authCodeRepo *repository.AuthCodeRepository,
	emailService interface {
		SendAuthCode(toEmail, code string) error
	},
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		authCodeRepo: authCodeRepo,
		emailService: emailService,
		jwtSecret:    jwtSecret,
	}
}

// RequestCodeResult результат запроса кода
type RequestCodeResult struct {
	IsNewUser bool
	ExpiresIn int // секунд
}

// RequestCode запрашивает код для входа/регистрации
func (s *AuthService) RequestCode(email string) (*RequestCodeResult, error) {
	// Валидация email
	if err := validators.Email.Validate(email); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	var user *models.User
	var isNewUser bool

	// Проверяем существование пользователя
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, 500, "database error")
		}
		// Пользователь не найден - создаём нового
		user = &models.User{
			Email:  email,
			Status: models.UserStatusPendingEmail,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, apperrors.Wrap(err, 500, "failed to create user")
		}
		isNewUser = true
	} else {
		user = existingUser
		isNewUser = false
	}

	// Проверяем rate limiting
	recentCodes, err := s.authCodeRepo.CountRecentByUserID(user.ID, codeRateLimitMinutes)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "database error")
	}
	if recentCodes >= maxCodesPerHour {
		return nil, apperrors.BadRequest("too many code requests, please try again later")
	}

	// Инвалидируем предыдущие коды
	_ = s.authCodeRepo.InvalidateAllForUser(user.ID)

	// Генерируем новый код
	code := generateAuthCode()
	authCode := &models.AuthCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(codeValidityMinutes * time.Minute),
	}
	if err := s.authCodeRepo.Create(authCode); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create auth code")
	}

	// Отправляем код на email
	if err := s.emailService.SendAuthCode(email, code); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to send email")
	}

	return &RequestCodeResult{
		IsNewUser: isNewUser,
		ExpiresIn: codeValidityMinutes * 60,
	}, nil
}

// AuthResult результат аутентификации
type AuthResult struct {
	Token    string
	User     *models.User
	NextStep string // "complete_profile" или пусто
}

// VerifyCode проверяет код и возвращает JWT
func (s *AuthService) VerifyCode(email, code string) (*AuthResult, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("invalid email or code")
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// Получаем валидный код
	authCode, err := s.authCodeRepo.GetValidCode(user.ID, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Увеличиваем счётчик попыток для последнего кода
			if latestCode, _ := s.authCodeRepo.GetLatestByUserID(user.ID); latestCode != nil {
				_ = s.authCodeRepo.IncrementAttempts(latestCode.ID)
			}
			return nil, apperrors.BadRequest("invalid or expired code")
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// Помечаем код как использованный
	if err := s.authCodeRepo.MarkAsUsed(authCode.ID); err != nil {
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// Обновляем статус пользователя
	var nextStep string
	if user.Status == models.UserStatusPendingEmail {
		// Подтверждаем email и переводим в pending_profile
		user.Status = models.UserStatusPendingProfile
		user.IsEmailVerified = true
		if err := s.userRepo.Update(user); err != nil {
			return nil, apperrors.Wrap(err, 500, "failed to update user")
		}
		nextStep = "complete_profile"
	} else if user.Status == models.UserStatusPendingProfile {
		nextStep = "complete_profile"
	}

	// Генерируем JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to generate token")
	}

	return &AuthResult{
		Token:    token,
		User:     user,
		NextStep: nextStep,
	}, nil
}

// CompleteProfile заполняет профиль пользователя
func (s *AuthService) CompleteProfile(userID int64, firstName, lastName string, middleName *string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем статус
	if user.Status != models.UserStatusPendingProfile {
		if user.Status == models.UserStatusActive {
			return apperrors.BadRequest("profile already completed")
		}
		return apperrors.BadRequest("email not verified")
	}

	// Валидация имени и фамилии
	if len(firstName) < 2 {
		return apperrors.BadRequest("first name must be at least 2 characters")
	}
	if len(lastName) < 2 {
		return apperrors.BadRequest("last name must be at least 2 characters")
	}

	// Обновляем профиль
	user.FirstName = &firstName
	user.LastName = &lastName
	user.MiddleName = middleName
	user.Status = models.UserStatusActive

	if err := s.userRepo.Update(user); err != nil {
		return apperrors.Wrap(err, 500, "failed to update profile")
	}

	return nil
}

// SetPassword устанавливает пароль пользователя
func (s *AuthService) SetPassword(userID int64, password string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем статус
	if user.Status != models.UserStatusActive {
		return apperrors.BadRequest("complete profile first")
	}

	// Валидация пароля
	if err := validators.Password.Validate(password); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to hash password")
	}

	// Сохраняем пароль
	passwordHash := string(hashedPassword)
	user.PasswordHash = &passwordHash
	if err := s.userRepo.Update(user); err != nil {
		return apperrors.Wrap(err, 500, "failed to save password")
	}

	return nil
}

// LoginWithPassword выполняет вход по паролю (если установлен)
func (s *AuthService) LoginWithPassword(email, password string) (*AuthResult, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем, установлен ли пароль
	if !user.HasPassword() {
		return nil, apperrors.BadRequest("password not set, use email code login")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Проверяем статус
	var nextStep string
	if user.Status == models.UserStatusPendingProfile {
		nextStep = "complete_profile"
	} else if user.Status == models.UserStatusPendingEmail {
		return nil, apperrors.BadRequest("email not verified")
	}

	// Генерируем JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to generate token")
	}

	return &AuthResult{
		Token:    token,
		User:     user,
		NextStep: nextStep,
	}, nil
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(userID int64) (*models.User, error) {
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}
	return user, nil
}

// generateJWT генерирует JWT токен
func (s *AuthService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":    user.ID,
		"email":  user.Email,
		"status": string(user.Status),
		"exp":    time.Now().Add(time.Hour * jwtExpirationHours).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateAuthCode генерирует 4-символьный код
// Исключены похожие символы: 0, O, 1, I, L
func generateAuthCode() string {
	const chars = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	code := make([]byte, 4)
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	for i := range code {
		code[i] = chars[int(randomBytes[i])%len(chars)]
	}
	return string(code)
}

// CleanupExpiredCodes удаляет истёкшие коды
func (s *AuthService) CleanupExpiredCodes() error {
	return s.authCodeRepo.DeleteExpired()
}

// ResendCode повторно отправляет код
func (s *AuthService) ResendCode(email string) (*RequestCodeResult, error) {
	return s.RequestCode(email)
}

// RefreshToken обновляет токен пользователя (получает новый по userID)
func (s *AuthService) RefreshToken(userID int64) (string, error) {
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.ErrUserNotFound
		}
		return "", apperrors.Wrap(err, 500, "database error")
	}
	return s.generateJWT(user)
}

// ChangePassword меняет пароль пользователя
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Если пароль установлен, проверяем старый
	if user.HasPassword() {
		if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword)); err != nil {
			return apperrors.BadRequest("invalid old password")
		}
	}

	// Валидация нового пароля
	if err := validators.Password.Validate(newPassword); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Хешируем и сохраняем
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to hash password")
	}

	passwordHash := string(hashedPassword)
	user.PasswordHash = &passwordHash
	if err := s.userRepo.Update(user); err != nil {
		return apperrors.Wrap(err, 500, "failed to save password")
	}

	return nil
}

// RemovePassword удаляет пароль пользователя (вход только через код)
func (s *AuthService) RemovePassword(userID int64, currentPassword string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем текущий пароль
	if !user.HasPassword() {
		return apperrors.BadRequest("password not set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(currentPassword)); err != nil {
		return apperrors.BadRequest("invalid password")
	}

	// Удаляем пароль
	user.PasswordHash = nil
	if err := s.userRepo.Update(user); err != nil {
		return apperrors.Wrap(err, 500, "failed to remove password")
	}

	return nil
}

// Deprecated: для совместимости, используйте AuthService
type legacyMethods interface{}

var _ legacyMethods = (*AuthService)(nil)

// Добавим алиас для обратной совместимости
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
	return s.generateJWT(user)
}

package services

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/osi-oss/osi/internal/logger"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret-key-for-jwt-testing"

// mockEmailService мок для email сервиса
type mockEmailService struct {
	sentCodes map[string]string // email -> code
	sendError error
}

func newMockEmailService() *mockEmailService {
	return &mockEmailService{
		sentCodes: make(map[string]string),
	}
}

func (m *mockEmailService) SendAuthCode(toEmail, code string) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentCodes[toEmail] = code
	return nil
}

// setupAuthTestDB создает тестовую базу данных для auth тестов
func setupAuthTestDB(t *testing.T) *gorm.DB {
	// Инициализируем logger для тестов
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger.Log = slog.New(handler)
	slog.SetDefault(logger.Log)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(
		&models.User{},
		&models.AuthCode{},
	)
	require.NoError(t, err, "failed to run migrations")

	return db
}

// createAuthService создает AuthService с моками
func createAuthService(t *testing.T, db *gorm.DB) (*AuthService, *mockEmailService) {
	userRepo := repository.NewUserRepository(db)
	authCodeRepo := repository.NewAuthCodeRepository(db)
	emailService := newMockEmailService()

	service := NewAuthService(userRepo, authCodeRepo, emailService, testJWTSecret)
	return service, emailService
}

// TestRequestCode_NewUser тестирует запрос кода для нового пользователя
func TestRequestCode_NewUser(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	email := "newuser@example.com"
	result, err := service.RequestCode(email)

	require.NoError(t, err)
	assert.True(t, result.IsNewUser)
	assert.Equal(t, codeValidityMinutes*60, result.ExpiresIn)

	// Проверяем что код отправлен
	code, ok := emailService.sentCodes[email]
	assert.True(t, ok, "code should be sent to email")
	assert.Len(t, code, 4, "code should be 4 characters")

	// Проверяем что пользователь создан
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	require.NoError(t, err)
	assert.Equal(t, models.UserStatusPendingEmail, user.Status)
}

// TestRequestCode_ExistingUser тестирует запрос кода для существующего пользователя
func TestRequestCode_ExistingUser(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	// Создаем существующего пользователя
	email := "existing@example.com"
	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	result, err := service.RequestCode(email)

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)

	// Проверяем что код отправлен
	_, ok := emailService.sentCodes[email]
	assert.True(t, ok)
}

// TestRequestCode_InvalidEmail тестирует запрос с невалидным email
func TestRequestCode_InvalidEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	_, err := service.RequestCode("invalid-email")
	assert.Error(t, err)
}

// TestRequestCode_RateLimit тестирует rate limiting
func TestRequestCode_RateLimit(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "ratelimit@example.com"

	// Создаем пользователя и много кодов
	user := &models.User{Email: email, Status: models.UserStatusActive}
	require.NoError(t, db.Create(user).Error)

	for i := 0; i < maxCodesPerHour; i++ {
		authCode := &models.AuthCode{
			UserID:    user.ID,
			Code:      "TEST",
			ExpiresAt: time.Now().Add(20 * time.Minute),
		}
		require.NoError(t, db.Create(authCode).Error)
	}

	// Следующий запрос должен вернуть ошибку
	_, err := service.RequestCode(email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many code requests")
}

// TestVerifyCode_Success тестирует успешную верификацию кода
func TestVerifyCode_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	email := "verify@example.com"

	// Запрашиваем код
	_, err := service.RequestCode(email)
	require.NoError(t, err)

	code := emailService.sentCodes[email]

	// Верифицируем код
	result, err := service.VerifyCode(email, code)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Token)
	assert.Equal(t, email, result.User.Email)
	assert.Equal(t, "complete_profile", result.NextStep)
	assert.Equal(t, models.UserStatusPendingProfile, result.User.Status)
}

// TestVerifyCode_InvalidCode тестирует неверный код
func TestVerifyCode_InvalidCode(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "verify@example.com"
	_, err := service.RequestCode(email)
	require.NoError(t, err)

	_, err = service.VerifyCode(email, "XXXX")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired code")
}

// TestVerifyCode_ExpiredCode тестирует истёкший код
func TestVerifyCode_ExpiredCode(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "expired@example.com"
	user := &models.User{Email: email, Status: models.UserStatusPendingEmail}
	require.NoError(t, db.Create(user).Error)

	// Создаём истёкший код
	expiredCode := &models.AuthCode{
		UserID:    user.ID,
		Code:      "TEST",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // истёк час назад
	}
	require.NoError(t, db.Create(expiredCode).Error)

	_, err := service.VerifyCode(email, "TEST")
	assert.Error(t, err)
}

// TestCompleteProfile_Success тестирует успешное заполнение профиля
func TestCompleteProfile_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	email := "profile@example.com"

	// Регистрируемся и верифицируем код
	_, err := service.RequestCode(email)
	require.NoError(t, err)

	code := emailService.sentCodes[email]
	result, err := service.VerifyCode(email, code)
	require.NoError(t, err)

	// Заполняем профиль
	middleName := "Иванович"
	profileResult, err := service.CompleteProfile(result.User.ID, "Иван", "Петров", &middleName)
	require.NoError(t, err)
	assert.NotNil(t, profileResult)
	assert.NotEmpty(t, profileResult.Token)
	assert.NotNil(t, profileResult.User)

	// Проверяем результат
	var user models.User
	require.NoError(t, db.First(&user, result.User.ID).Error)
	assert.Equal(t, "Иван", *user.FirstName)
	assert.Equal(t, "Петров", *user.LastName)
	assert.Equal(t, "Иванович", *user.MiddleName)
	assert.Equal(t, models.UserStatusActive, user.Status)
}

// TestCompleteProfile_ShortName тестирует короткое имя
func TestCompleteProfile_ShortName(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "short@example.com",
		Status: models.UserStatusPendingProfile,
	}
	require.NoError(t, db.Create(user).Error)

	_, err := service.CompleteProfile(user.ID, "A", "B", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 characters")
}

// TestCompleteProfile_WrongStatus тестирует неверный статус
func TestCompleteProfile_WrongStatus(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	// Пользователь уже активен
	user := &models.User{
		Email:  "active@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	_, err := service.CompleteProfile(user.ID, "Иван", "Петров", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already completed")
}

// TestSetPassword_Success тестирует установку пароля
func TestSetPassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "password@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	err := service.SetPassword(user.ID, "SecurePass123")
	require.NoError(t, err)

	// Проверяем что пароль установлен
	var updatedUser models.User
	require.NoError(t, db.First(&updatedUser, user.ID).Error)
	assert.NotNil(t, updatedUser.PasswordHash)
	assert.True(t, updatedUser.HasPassword())
}

// TestSetPassword_WeakPassword тестирует слабый пароль
func TestSetPassword_WeakPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "weak@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	err := service.SetPassword(user.ID, "123")
	assert.Error(t, err)
}

// TestSetPassword_NotActive тестирует установку пароля для неактивного пользователя
func TestSetPassword_NotActive(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "pending@example.com",
		Status: models.UserStatusPendingProfile,
	}
	require.NoError(t, db.Create(user).Error)

	err := service.SetPassword(user.ID, "SecurePass123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "complete profile first")
}

// TestLoginWithPassword_Success тестирует вход по паролю
func TestLoginWithPassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "login@example.com"
	password := "SecurePass123"

	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, password))

	result, err := service.LoginWithPassword(email, password)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, email, result.User.Email)
	assert.Empty(t, result.NextStep)
}

// TestLoginWithPassword_WrongPassword тестирует неверный пароль
func TestLoginWithPassword_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "wrongpass@example.com"

	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, "CorrectPass123"))

	_, err := service.LoginWithPassword(email, "WrongPass456")
	assert.Error(t, err)
}

// TestLoginWithPassword_NoPassword тестирует вход без пароля
func TestLoginWithPassword_NoPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "nopassword@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	_, err := service.LoginWithPassword(user.Email, "SomePass123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password not set")
}

// TestChangePassword_Success тестирует смену пароля
func TestChangePassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	email := "change@example.com"
	oldPassword := "OldPass123"
	newPassword := "NewPass456"

	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, oldPassword))

	err := service.ChangePassword(user.ID, oldPassword, newPassword)
	require.NoError(t, err)

	// Проверяем что новый пароль работает
	_, err = service.LoginWithPassword(email, newPassword)
	require.NoError(t, err)

	// Старый пароль не работает
	_, err = service.LoginWithPassword(email, oldPassword)
	assert.Error(t, err)
}

// TestChangePassword_WrongOldPassword тестирует неверный старый пароль
func TestChangePassword_WrongOldPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "wrongold@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, "CorrectOld123"))

	err := service.ChangePassword(user.ID, "WrongOld123", "NewPass456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid old password")
}

// TestRemovePassword_Success тестирует удаление пароля
func TestRemovePassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	password := "SecurePass123"

	user := &models.User{
		Email:  "remove@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, password))

	err := service.RemovePassword(user.ID, password)
	require.NoError(t, err)

	// Проверяем что пароль удалён
	var updatedUser models.User
	require.NoError(t, db.First(&updatedUser, user.ID).Error)
	assert.False(t, updatedUser.HasPassword())
}

// TestRemovePassword_WrongPassword тестирует удаление с неверным паролем
func TestRemovePassword_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "removewrong@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, service.SetPassword(user.ID, "CorrectPass123"))

	err := service.RemovePassword(user.ID, "WrongPass456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

// TestGenerateJWT тестирует генерацию JWT
func TestGenerateJWT(t *testing.T) {
	db := setupAuthTestDB(t)
	service, _ := createAuthService(t, db)

	user := &models.User{
		Email:  "jwt@example.com",
		Status: models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	token, err := service.GenerateJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestFullRegistrationFlow тестирует полный флоу регистрации
func TestFullRegistrationFlow(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	email := "fullflow@example.com"

	// 1. Запрос кода
	reqResult, err := service.RequestCode(email)
	require.NoError(t, err)
	assert.True(t, reqResult.IsNewUser)

	// 2. Верификация кода
	code := emailService.sentCodes[email]
	verifyResult, err := service.VerifyCode(email, code)
	require.NoError(t, err)
	assert.Equal(t, "complete_profile", verifyResult.NextStep)

	// 3. Заполнение профиля
	profileResult, err := service.CompleteProfile(verifyResult.User.ID, "Алексей", "Смирнов", nil)
	require.NoError(t, err)
	assert.NotNil(t, profileResult)
	assert.NotEmpty(t, profileResult.Token)

	// 4. Установка пароля (опционально)
	err = service.SetPassword(verifyResult.User.ID, "MySecurePass123")
	require.NoError(t, err)

	// 5. Вход по паролю
	loginResult, err := service.LoginWithPassword(email, "MySecurePass123")
	require.NoError(t, err)
	assert.NotEmpty(t, loginResult.Token)
	assert.Equal(t, models.UserStatusActive, loginResult.User.Status)
}

// TestFullLoginFlow тестирует полный флоу входа
func TestFullLoginFlow(t *testing.T) {
	db := setupAuthTestDB(t)
	service, emailService := createAuthService(t, db)

	// Создаём активного пользователя
	email := "loginflow@example.com"
	firstName := "Тест"
	lastName := "Юзер"
	user := &models.User{
		Email:     email,
		FirstName: &firstName,
		LastName:  &lastName,
		Status:    models.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	// 1. Запрос кода
	reqResult, err := service.RequestCode(email)
	require.NoError(t, err)
	assert.False(t, reqResult.IsNewUser)

	// 2. Верификация кода
	code := emailService.sentCodes[email]
	verifyResult, err := service.VerifyCode(email, code)
	require.NoError(t, err)
	assert.Empty(t, verifyResult.NextStep) // уже активный
	assert.NotEmpty(t, verifyResult.Token)
}

// TestGenerateAuthCode тестирует генерацию кода
func TestGenerateAuthCode(t *testing.T) {
	codes := make(map[string]bool)

	// Генерируем 100 кодов и проверяем уникальность
	for i := 0; i < 100; i++ {
		code := generateAuthCode()
		assert.Len(t, code, 4)

		// Проверяем что нет похожих символов
		for _, c := range code {
			assert.NotContains(t, "01ILO", string(c))
		}

		codes[code] = true
	}

	// Должно быть много уникальных кодов (почти все)
	assert.Greater(t, len(codes), 90)
}

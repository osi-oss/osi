package services

import (
	"testing"
	"time"

	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// stubEmailSvc реализует минимально нужный интерфейс EmailService для тестов
type stubEmailSvc struct{}

func (s *stubEmailSvc) SendPasswordResetEmail(toEmail, token, baseURL string) error { return nil }
func (s *stubEmailSvc) SendWelcomeEmail(toEmail, userName string) error             { return nil }

// setupTestDB создает in-memory sqlite и прогоняет миграции
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&models.User{}, &models.PasswordResetToken{}))
	return db
}

func setupTestEnv(t *testing.T) (*gorm.DB, *UserService) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	resetRepo := repository.NewPasswordResetRepository(db)

	svc := NewUserService(
		userRepo,
		resetRepo,
		&stubEmailSvc{},
		"test-secret",
		"http://localhost::3030")

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	return db, svc
}

func Test_SignUp_Simple(t *testing.T) {
	_, svc := setupTestEnv(t)

	user, err := svc.SignUp("test@example.com", "abcD1234")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
}
func Test_SignUp(t *testing.T) {
	_, svc := setupTestEnv(t)
	user, err := svc.SignUp("test@example.com", "Abcd1234")
	require.NoError(t, err)
	require.Equal(t, "test@example.com", user.Email)

	// повторная регистрация должна выдать ошибку
	_, err = svc.SignUp("test@example.com", "Abcd1234")
	require.Error(t, err)
}

func Test_LogIn_Simple(t *testing.T) {
	_, svc := setupTestEnv(t)

	// регистрируем пользователя
	_, err := svc.SignUp("test@example.com", "Abcd1234")
	require.NoError(t, err)

	// правильный логин
	token, err := svc.LogIn("test@example.com", "Abcd1234")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}
func Test_LogIn(t *testing.T) {
	_, svc := setupTestEnv(t)

	// регистрируем пользователя
	_, err := svc.SignUp("test@example.com", "Abcd1234")
	require.NoError(t, err)

	// правильный логин
	token, err := svc.LogIn("test@example.com", "Abcd1234")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// неправильный пароль
	_, err = svc.LogIn("test@example.com", "wrongpass")
	require.Error(t, err)
}

func TestUserService_SignUp_And_LogIn(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	resetRepo := repository.NewPasswordResetRepository(db)

	// stubbed email service to avoid real SMTP calls
	svc := NewUserService(userRepo, resetRepo, &stubEmailSvc{}, "test-secret", "http://localhost")

	// Sign up
	user, err := svc.SignUp("test@example.com", "Abcd1234")
	require.NoError(t, err)
	require.Equal(t, "test@example.com", user.Email)

	// Log in should return a token
	token, err := svc.LogIn("test@example.com", "Abcd1234")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// invalid password
	_, err = svc.LogIn("test@example.com", "wrongpass")
	require.Error(t, err)

	// Test password reset flow: request and validate
	require.NoError(t, svc.RequestPasswordReset("test@example.com"))

	// Create a manual token to test ResetPassword
	tokenStr := "manualtoken123"
	reset := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}
	require.NoError(t, resetRepo.Create(reset))

	// Reset password with validator-compliant password
	require.NoError(t, svc.ResetPassword(tokenStr, "Newpass123"))
}

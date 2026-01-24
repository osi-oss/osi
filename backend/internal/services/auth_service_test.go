package services

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/osi-oss/osi/internal/logger"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret-key-for-jwt-testing"

// mockEmailService мок для email сервиса
type mockEmailService struct {
	sentCodes map[string]string
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

// AuthServiceTestSuite - тестовый набор для AuthService
type AuthServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	service      *AuthService
	emailService *mockEmailService
	userRepo     *repository.UserRepository
	authCodeRepo *repository.AuthCodeRepository
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *AuthServiceTestSuite) SetupSuite() {
	// Инициализируем logger для тестов
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Меньше логов в тестах
	})
	logger.Log = slog.New(handler)
	slog.SetDefault(logger.Log)
}

// SetupTest выполняется перед каждым тестом
func (s *AuthServiceTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err, "failed to open test database")

	err = s.db.AutoMigrate(
		&models.User{},
		&models.AuthCode{},
	)
	s.Require().NoError(err, "failed to run migrations")

	s.userRepo = repository.NewUserRepository(s.db)
	s.authCodeRepo = repository.NewAuthCodeRepository(s.db)
	s.emailService = newMockEmailService()

	s.service = NewAuthService(s.userRepo, s.authCodeRepo, s.emailService, testJWTSecret)
}

// TearDownTest выполняется после каждого теста
func (s *AuthServiceTestSuite) TearDownTest() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// TestRequestCode_NewUser тестирует запрос кода для нового пользователя
func (s *AuthServiceTestSuite) TestRequestCode_NewUser() {
	email := "newuser@example.com"
	result, err := s.service.RequestCode(email)

	s.Require().NoError(err)
	s.Assert().True(result.IsNewUser)
	s.Assert().Equal(codeValidityMinutes*60, result.ExpiresIn)

	// Проверяем что код отправлен
	code, ok := s.emailService.sentCodes[email]
	s.Assert().True(ok, "code should be sent to email")
	s.Assert().Len(code, 4, "code should be 4 characters")

	// Проверяем что пользователь создан
	var user models.User
	err = s.db.Where("email = ?", email).First(&user).Error
	s.Require().NoError(err)
	s.Assert().Equal(models.UserStatusPendingEmail, user.Status)
}

// TestRequestCode_ExistingUser тестирует запрос кода для существующего пользователя
func (s *AuthServiceTestSuite) TestRequestCode_ExistingUser() {
	email := "existing@example.com"
	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(user).Error)

	result, err := s.service.RequestCode(email)

	s.Require().NoError(err)
	s.Assert().False(result.IsNewUser)

	// Проверяем что код отправлен
	_, ok := s.emailService.sentCodes[email]
	s.Assert().True(ok)
}

// TestRequestCode_InvalidEmail тестирует запрос с невалидным email
func (s *AuthServiceTestSuite) TestRequestCode_InvalidEmail() {
	testCases := []struct {
		name  string
		email string
	}{
		{"without @", "invalid-email"},
		{"without domain", "test@"},
		{"empty", ""},
		{"only @", "@"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.service.RequestCode(tc.email)
			s.Assert().Error(err)
		})
	}
}

// TestRequestCode_RateLimit тестирует rate limiting
func (s *AuthServiceTestSuite) TestRequestCode_RateLimit() {
	email := "ratelimit@example.com"

	// Создаём пользователя и много кодов
	user := &models.User{Email: email, Status: models.UserStatusActive}
	s.Require().NoError(s.db.Create(user).Error)

	for i := 0; i < maxCodesPerHour; i++ {
		authCode := &models.AuthCode{
			UserID:    user.ID,
			Code:      "TEST",
			ExpiresAt: time.Now().Add(20 * time.Minute),
		}
		s.Require().NoError(s.db.Create(authCode).Error)
	}

	// Следующий запрос должен вернуть ошибку
	_, err := s.service.RequestCode(email)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "too many code requests")
}

// TestVerifyCode_Success тестирует успешную верификацию кода
func (s *AuthServiceTestSuite) TestVerifyCode_Success() {
	email := "verify@example.com"

	// Запрашиваем код
	_, err := s.service.RequestCode(email)
	s.Require().NoError(err)

	code := s.emailService.sentCodes[email]

	// Верифицируем код
	result, err := s.service.VerifyCode(email, code)
	s.Require().NoError(err)

	s.Assert().NotEmpty(result.Token)
	s.Assert().Equal(email, result.User.Email)
	s.Assert().Equal("complete_profile", result.NextStep)
	s.Assert().Equal(models.UserStatusPendingProfile, result.User.Status)
}

// TestVerifyCode_InvalidCode тестирует неверный код
func (s *AuthServiceTestSuite) TestVerifyCode_InvalidCode() {
	email := "verify@example.com"
	_, err := s.service.RequestCode(email)
	s.Require().NoError(err)

	_, err = s.service.VerifyCode(email, "XXXX")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "invalid or expired code")
}

// TestVerifyCode_ExpiredCode тестирует истёкший код
func (s *AuthServiceTestSuite) TestVerifyCode_ExpiredCode() {
	email := "expired@example.com"
	user := &models.User{Email: email, Status: models.UserStatusPendingEmail}
	s.Require().NoError(s.db.Create(user).Error)

	// Создаём истёкший код
	expiredCode := &models.AuthCode{
		UserID:    user.ID,
		Code:      "TEST",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	s.Require().NoError(s.db.Create(expiredCode).Error)

	_, err := s.service.VerifyCode(email, "TEST")
	s.Assert().Error(err)
}

// TestCompleteProfile_Success тестирует успешное заполнение профиля
func (s *AuthServiceTestSuite) TestCompleteProfile_Success() {
	email := "profile@example.com"

	// Регистрируемся и верифицируем код
	_, err := s.service.RequestCode(email)
	s.Require().NoError(err)

	code := s.emailService.sentCodes[email]
	result, err := s.service.VerifyCode(email, code)
	s.Require().NoError(err)

	// Заполняем профиль
	middleName := "Иванович"
	profileResult, err := s.service.CompleteProfile(result.User.ID, "Иван", "Петров", &middleName)
	s.Require().NoError(err)
	s.Assert().NotNil(profileResult)
	s.Assert().NotEmpty(profileResult.Token)
	s.Assert().NotNil(profileResult.User)

	// Проверяем результат
	var user models.User
	s.Require().NoError(s.db.First(&user, result.User.ID).Error)
	s.Assert().Equal("Иван", *user.FirstName)
	s.Assert().Equal("Петров", *user.LastName)
	s.Assert().Equal("Иванович", *user.MiddleName)
	s.Assert().Equal(models.UserStatusActive, user.Status)
}

// TestCompleteProfile_Validation тестирует валидацию при заполнении профиля
func (s *AuthServiceTestSuite) TestCompleteProfile_Validation() {
	user := &models.User{
		Email:  "test@example.com",
		Status: models.UserStatusPendingProfile,
	}
	s.Require().NoError(s.db.Create(user).Error)

	testCases := []struct {
		name       string
		firstName  string
		lastName   string
		middleName *string
		expectErr  bool
		errMsg     string
	}{
		{
			name:      "short first name",
			firstName: "A",
			lastName:  "Valid",
			expectErr: true,
			errMsg:    "at least 2 characters",
		},
		{
			name:      "short last name",
			firstName: "Valid",
			lastName:  "B",
			expectErr: true,
			errMsg:    "at least 2 characters",
		},
		{
			name:      "valid names",
			firstName: "Иван",
			lastName:  "Петров",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.service.CompleteProfile(user.ID, tc.firstName, tc.lastName, tc.middleName)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Contains(err.Error(), tc.errMsg)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

// TestPassword_FullCycle тестирует полный цикл работы с паролем
func (s *AuthServiceTestSuite) TestPassword_FullCycle() {
	email := "password@example.com"
	password := "SecurePass123"
	newPassword := "NewSecurePass456"

	user := &models.User{
		Email:  email,
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(user).Error)

	// 1. Установка пароля
	err := s.service.SetPassword(user.ID, password)
	s.Require().NoError(err)

	var updatedUser models.User
	s.Require().NoError(s.db.First(&updatedUser, user.ID).Error)
	s.Assert().True(updatedUser.HasPassword())

	// 2. Вход с паролем
	loginResult, err := s.service.LoginWithPassword(email, password)
	s.Require().NoError(err)
	s.Assert().NotEmpty(loginResult.Token)
	s.Assert().Equal(email, loginResult.User.Email)

	// 3. Вход с неправильным паролем
	_, err = s.service.LoginWithPassword(email, "WrongPass")
	s.Assert().Error(err)

	// 4. Смена пароля
	err = s.service.ChangePassword(user.ID, password, newPassword)
	s.Require().NoError(err)

	// 5. Вход с новым паролем работает
	_, err = s.service.LoginWithPassword(email, newPassword)
	s.Assert().NoError(err)

	// 6. Вход со старым паролем не работает
	_, err = s.service.LoginWithPassword(email, password)
	s.Assert().Error(err)

	// 7. Удаление пароля
	err = s.service.RemovePassword(user.ID, newPassword)
	s.Require().NoError(err)

	s.Require().NoError(s.db.First(&updatedUser, user.ID).Error)
	s.Assert().False(updatedUser.HasPassword())

	// 8. Вход с паролем больше не работает
	_, err = s.service.LoginWithPassword(email, newPassword)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "password not set")
}

// TestFullRegistrationFlow тестирует полный флоу регистрации
func (s *AuthServiceTestSuite) TestFullRegistrationFlow() {
	email := "fullflow@example.com"

	// 1. Запрос кода
	reqResult, err := s.service.RequestCode(email)
	s.Require().NoError(err)
	s.Assert().True(reqResult.IsNewUser)

	// 2. Верификация кода
	code := s.emailService.sentCodes[email]
	verifyResult, err := s.service.VerifyCode(email, code)
	s.Require().NoError(err)
	s.Assert().Equal("complete_profile", verifyResult.NextStep)

	// 3. Заполнение профиля
	profileResult, err := s.service.CompleteProfile(verifyResult.User.ID, "Алексей", "Смирнов", nil)
	s.Require().NoError(err)
	s.Assert().NotEmpty(profileResult.Token)

	// 4. Установка пароля (опционально)
	err = s.service.SetPassword(verifyResult.User.ID, "MySecurePass123")
	s.Require().NoError(err)

	// 5. Вход по паролю
	loginResult, err := s.service.LoginWithPassword(email, "MySecurePass123")
	s.Require().NoError(err)
	s.Assert().NotEmpty(loginResult.Token)
	s.Assert().Equal(models.UserStatusActive, loginResult.User.Status)
}

// TestFullLoginFlow тестирует полный флоу входа
func (s *AuthServiceTestSuite) TestFullLoginFlow() {
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
	s.Require().NoError(s.db.Create(user).Error)

	// 1. Запрос кода
	reqResult, err := s.service.RequestCode(email)
	s.Require().NoError(err)
	s.Assert().False(reqResult.IsNewUser)

	// 2. Верификация кода
	code := s.emailService.sentCodes[email]
	verifyResult, err := s.service.VerifyCode(email, code)
	s.Require().NoError(err)
	s.Assert().Empty(verifyResult.NextStep) // уже активный
	s.Assert().NotEmpty(verifyResult.Token)
}

// TestGenerateAuthCode тестирует генерацию кода
func (s *AuthServiceTestSuite) TestGenerateAuthCode() {
	codes := make(map[string]bool)

	// Генерируем 100 кодов и проверяем уникальность
	for i := 0; i < 100; i++ {
		code := generateAuthCode()
		s.Assert().Len(code, 4)

		// Проверяем что нет похожих символов
		for _, c := range code {
			s.Assert().NotContains("01ILO", string(c))
		}

		codes[code] = true
	}

	// Должно быть много уникальных кодов (почти все)
	s.Assert().Greater(len(codes), 90)
}

// TestRefreshToken тестирует обновление токена
func (s *AuthServiceTestSuite) TestRefreshToken() {
	s.T().Skip("Temporarily disabled")
	user := &models.User{
		Email:  "refresh@example.com",
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(user).Error)

	// Генерируем токен
	token1, err := s.service.RefreshToken(user.ID)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token1)

	// Генерируем ещё один токен
	token2, err := s.service.RefreshToken(user.ID)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token2)

	// Оба токена валидны, но разные
	s.Assert().NotEqual(token1, token2)
}

// TestGetUserByID тестирует получение пользователя по ID
func (s *AuthServiceTestSuite) TestGetUserByID() {
	user := &models.User{
		Email:  "getuser@example.com",
		Status: models.UserStatusActive,
	}
	s.Require().NoError(s.db.Create(user).Error)

	// Успешное получение
	foundUser, err := s.service.GetUserByID(user.ID)
	s.Require().NoError(err)
	s.Assert().Equal(user.Email, foundUser.Email)

	// Пользователь не найден
	_, err = s.service.GetUserByID(99999)
	s.Assert().Error(err)
}

// Запуск тестового набора
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

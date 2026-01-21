package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// RequestCode godoc
// @Summary      Запрос кода для входа/регистрации
// @Description  Отправляет 4-символьный код на указанный email.
// @Description  Если пользователь не существует, создаётся новый аккаунт.
// @Description  Код действителен 20 минут. Максимум 5 запросов в час.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.RequestCodeRequest true "Email для получения кода"
// @Success      200 {object} dto.RequestCodeResponse "Код отправлен"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат email или превышен лимит запросов"
// @Failure      500 {object} dto.ErrorResponse "Ошибка отправки email"
// @Router       /auth/request-code [post]
func (ctrl *AuthController) RequestCode(c *gin.Context) {
	var req dto.RequestCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctrl.authService.RequestCode(req.Email)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.RequestCodeResponse{
		Message:   "Verification code sent to your email",
		IsNewUser: result.IsNewUser,
		ExpiresIn: result.ExpiresIn,
	})
}

// VerifyCode godoc
// @Summary      Проверка кода и получение токена
// @Description  Проверяет код подтверждения и возвращает JWT токен.
// @Description  Для новых пользователей требуется заполнение профиля (next_step: complete_profile).
// @Description  Максимум 5 попыток ввода кода.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyCodeRequest true "Email и код подтверждения"
// @Success      200 {object} dto.AuthResponse "Успешная аутентификация"
// @Failure      400 {object} dto.ErrorResponse "Неверный или истёкший код"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /auth/verify-code [post]
func (ctrl *AuthController) VerifyCode(c *gin.Context) {
	var req dto.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctrl.authService.VerifyCode(req.Email, req.Code)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.AuthResponse{
		Token:     result.Token,
		ExpiresIn: 24 * 60 * 60, // 24 часа в секундах
		User:      dto.ToUserResponse(result.User),
		NextStep:  result.NextStep,
	})
}

// ResendCode godoc
// @Summary      Повторная отправка кода
// @Description  Повторно отправляет код подтверждения на email.
// @Description  Предыдущий код становится недействительным.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.RequestCodeRequest true "Email для получения кода"
// @Success      200 {object} dto.RequestCodeResponse "Код отправлен повторно"
// @Failure      400 {object} dto.ErrorResponse "Превышен лимит запросов"
// @Failure      500 {object} dto.ErrorResponse "Ошибка отправки email"
// @Router       /auth/resend-code [post]
func (ctrl *AuthController) ResendCode(c *gin.Context) {
	var req dto.RequestCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctrl.authService.ResendCode(req.Email)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.RequestCodeResponse{
		Message:   "Verification code resent to your email",
		IsNewUser: result.IsNewUser,
		ExpiresIn: result.ExpiresIn,
	})
}

// CompleteProfile godoc
// @Summary      Заполнение профиля
// @Description  Заполняет обязательные данные профиля после подтверждения email.
// @Description  Требуется для новых пользователей перед доступом к системе.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CompleteProfileRequest true "Данные профиля"
// @Success      200 {object} dto.ProfileResponse "Профиль успешно заполнен"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации или профиль уже заполнен"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /auth/complete-profile [post]
func (ctrl *AuthController) CompleteProfile(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	var req dto.CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.authService.CompleteProfile(userID, req.FirstName, req.LastName, req.MiddleName); err != nil {
		helpers.RespondError(c, err)
		return
	}

	// Получаем обновлённого пользователя
	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ProfileResponse{
		Message: "Profile completed successfully",
		User:    dto.ToUserResponse(user),
	})
}

// LoginWithPassword godoc
// @Summary      Вход по паролю
// @Description  Авторизация по email и паролю (если пароль установлен).
// @Description  Если пароль не установлен, используйте /auth/request-code.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.PasswordLoginRequest true "Email и пароль"
// @Success      200 {object} dto.AuthResponse "Успешная авторизация"
// @Failure      400 {object} dto.ErrorResponse "Пароль не установлен"
// @Failure      401 {object} dto.ErrorResponse "Неверный email или пароль"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /auth/login-password [post]
func (ctrl *AuthController) LoginWithPassword(c *gin.Context) {
	var req dto.PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctrl.authService.LoginWithPassword(req.Email, req.Password)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.AuthResponse{
		Token:     result.Token,
		ExpiresIn: 24 * 60 * 60,
		User:      dto.ToUserResponse(result.User),
		NextStep:  result.NextStep,
	})
}

// SetPassword godoc
// @Summary      Установка пароля
// @Description  Устанавливает пароль для быстрого входа (опционально).
// @Description  Требуется статус active (профиль должен быть заполнен).
// @Tags         Профиль
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.SetPasswordRequest true "Новый пароль"
// @Success      200 {object} dto.MessageResponse "Пароль установлен"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации пароля"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /profile/set-password [post]
func (ctrl *AuthController) SetPassword(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	var req dto.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.authService.SetPassword(userID, req.Password); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "Password set successfully"})
}

// ChangePassword godoc
// @Summary      Изменение пароля
// @Description  Изменяет существующий пароль на новый.
// @Tags         Профиль
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ChangePasswordRequest true "Старый и новый пароль"
// @Success      200 {object} dto.MessageResponse "Пароль изменён"
// @Failure      400 {object} dto.ErrorResponse "Неверный старый пароль или ошибка валидации"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /profile/change-password [post]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "Password changed successfully"})
}

// RemovePassword godoc
// @Summary      Удаление пароля
// @Description  Удаляет пароль. После этого вход возможен только через код на email.
// @Tags         Профиль
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.RemovePasswordRequest true "Текущий пароль для подтверждения"
// @Success      200 {object} dto.MessageResponse "Пароль удалён"
// @Failure      400 {object} dto.ErrorResponse "Неверный пароль"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /profile/remove-password [post]
func (ctrl *AuthController) RemovePassword(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	var req dto.RemovePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.authService.RemovePassword(userID, req.Password); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "Password removed successfully"})
}

// GetProfile godoc
// @Summary      Получение профиля
// @Description  Возвращает данные текущего авторизованного пользователя.
// @Tags         Профиль
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.ProfileResponse "Данные профиля"
// @Failure      401 {object} dto.ErrorResponse "Требуется авторизация"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /profile [get]
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ProfileResponse{
		Message: "Profile retrieved successfully",
		User:    dto.ToUserResponse(user),
	})
}

// Logout godoc
// @Summary      Выход из системы
// @Description  Завершает сессию пользователя. Удаляет cookie с токеном.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.MessageResponse "Успешный выход"
// @Router       /auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	helpers.RespondOK(c, gin.H{"message": "Logged out successfully"})
}

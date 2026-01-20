package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// SignUp godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя в системе.
// @Description  Email должен быть уникальным, пароль минимум 6 символов.
// @Description  После успешной регистрации пользователь может авторизоваться через /login.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.SignUpRequest true "Данные для регистрации"
// @Success      201 {object} dto.SignUpResponse "Пользователь успешно создан"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: неверный формат email или пароль короче 6 символов"
// @Failure      409 {object} dto.ErrorResponse "Пользователь с таким email уже существует"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /signup [post]
func (ctrl *UserController) SignUp(c *gin.Context) {
	var req dto.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.userService.SignUp(req.Email, req.Password)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, gin.H{
		"message": "User created successfully",
		"user":    dto.ToUserResponse(user),
	})
}

// LogIn godoc
// @Summary      Авторизация пользователя
// @Description  Авторизует пользователя по email и паролю.
// @Description  Возвращает JWT токен, который нужно передавать в заголовке Authorization: Bearer <token>.
// @Description  Токен действителен 24 часа.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Учётные данные для входа"
// @Success      200 {object} dto.LoginResponse "Успешная авторизация, возвращается JWT токен"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации: неверный формат запроса"
// @Failure      401 {object} dto.ErrorResponse "Неверный email или пароль"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /login [post]
func (ctrl *UserController) LogIn(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ctrl.userService.LogIn(req.Email, req.Password)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

// GetProfile godoc
// @Summary      Получение профиля текущего пользователя
// @Description  Возвращает данные профиля авторизованного пользователя.
// @Description  Требует авторизацию через JWT токен.
// @Tags         Пользователь
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.ProfileResponse "Данные профиля пользователя"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /profile [get]
func (ctrl *UserController) GetProfile(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	user, err := ctrl.userService.GetByID(userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{
		"message": "Profile retrieved successfully",
		"user":    dto.ToUserResponse(user),
	})
}

// Logout godoc
// @Summary      Выход из системы
// @Description  Завершает сессию пользователя.
// @Description  Удаляет токен из cookies (auth_token).
// @Description  После выхода JWT токен остаётся валидным до истечения срока.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.MessageResponse "Успешный выход из системы"
// @Router       /logout [post]
func (ctrl *UserController) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	helpers.RespondOK(c, gin.H{"message": "Logged out successfully"})
}

// RequestPasswordReset godoc
// @Summary      Запрос сброса пароля
// @Description  Отправляет письмо с инструкциями по сбросу пароля на указанный email.
// @Description  В целях безопасности всегда возвращает успех, даже если email не найден.
// @Description  Токен сброса действителен 1 час.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.PasswordResetRequest true "Email для восстановления доступа"
// @Success      200 {object} dto.MessageResponse "Инструкции отправлены (если email существует)"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат email"
// @Router       /forgot-password [post]
func (ctrl *UserController) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.userService.RequestPasswordReset(req.Email); err != nil {
		log.Printf("Password reset error: %v", err)
	}

	// Всегда возвращаем успех (безопасность)
	helpers.RespondOK(c, gin.H{
		"message": "If email exists, password reset instructions have been sent",
	})
}

// ResetPassword godoc
// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль с использованием токена из письма.
// @Description  Токен можно использовать только один раз.
// @Description  Новый пароль должен быть не менее 6 символов.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Токен сброса и новый пароль"
// @Success      200 {object} dto.MessageResponse "Пароль успешно изменён"
// @Failure      400 {object} dto.ErrorResponse "Невалидный или истёкший токен, или пароль слишком короткий"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /reset-password [post]
func (ctrl *UserController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.userService.ResetPassword(req.Token, req.NewPassword); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, gin.H{"message": "Password has been reset successfully"})
}

// ValidateResetToken godoc
// @Summary      Проверка токена сброса пароля
// @Description  Проверяет, действителен ли токен для сброса пароля.
// @Description  Используется для предварительной проверки перед отображением формы сброса.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        token query string true "Токен из письма для сброса пароля" example(abc123def456)
// @Success      200 {object} dto.TokenValidationResponse "Токен валиден, можно сбрасывать пароль"
// @Failure      400 {object} dto.ErrorResponse "Токен не указан, невалиден или истёк"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /reset-password/validate [get]
func (ctrl *UserController) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	valid, err := ctrl.userService.ValidateResetToken(token)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}

	helpers.RespondOK(c, gin.H{
		"valid":   true,
		"message": "Token is valid",
	})
}

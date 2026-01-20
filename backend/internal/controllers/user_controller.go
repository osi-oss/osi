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

// SignUp регистрирует нового пользователя
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя с email и паролем
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.SignUpRequest true "Данные для регистрации"
// @Success      201  {object}  map[string]interface{}  "Пользователь создан"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      409  {object}  map[string]string  "Email уже существует"
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

// LogIn авторизует пользователя
// @Summary      Авторизация пользователя
// @Description  Авторизует пользователя и возвращает JWT токен
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Данные для входа"
// @Success      200  {object}  map[string]interface{}  "Успешная авторизация"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
// @Failure      401  {object}  map[string]string  "Неверные учетные данные"
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

// GetProfile возвращает профиль авторизованного пользователя
// @Summary      Получение профиля
// @Description  Возвращает профиль текущего авторизованного пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponse  "Профиль пользователя"
// @Failure      401  {object}  map[string]string  "Не авторизован"
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

// Logout выполняет выход пользователя
// @Summary      Выход из системы
// @Description  Удаляет токен авторизации
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Успешный выход"
// @Router       /logout [post]
func (ctrl *UserController) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	helpers.RespondOK(c, gin.H{"message": "Logged out successfully"})
}

// RequestPasswordReset запрашивает восстановление пароля
// @Summary      Запрос сброса пароля
// @Description  Отправляет email с инструкциями по сбросу пароля
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.PasswordResetRequest true "Email для сброса пароля"
// @Success      200  {object}  map[string]string  "Инструкции отправлены"
// @Failure      400  {object}  map[string]string  "Ошибка валидации"
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

// ResetPassword сбрасывает пароль по токену
// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль по токену из email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Токен и новый пароль"
// @Success      200  {object}  map[string]string  "Пароль успешно изменён"
// @Failure      400  {object}  map[string]string  "Невалидный токен"
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

// ValidateResetToken проверяет валидность токена
// @Summary      Проверка токена сброса пароля
// @Description  Проверяет, валиден ли токен для сброса пароля
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token query string true "Токен сброса пароля"
// @Success      200  {object}  map[string]interface{}  "Токен валиден"
// @Failure      400  {object}  map[string]string  "Токен невалиден или истёк"
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

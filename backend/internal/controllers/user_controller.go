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

// Logout удаляет токен из куки
func (ctrl *UserController) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	helpers.RespondOK(c, gin.H{"message": "Logged out successfully"})
}

// RequestPasswordReset запрашивает восстановление пароля
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

package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/services"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (ctrl *UserController) SignUp(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the. user
	user, err := ctrl.userService.SignUp(body.Email, body.Password)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			statusCode = http.StatusConflict
		case errors.Is(err, services.ErrInvalidInput):
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"verified":   user.Verified,
		"created_at": user.CreatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    response,
	})
}

// GetProfile возвращает профиль авторизованного пользователя
func (ctrl *UserController) GetProfile(c *gin.Context) {
	// Получаем userID из middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Конвертируем в uint
	id := uint(userID.(float64))

	user, err := ctrl.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Возвращаем профиль без пароля
	response := gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"verified":   user.Verified,
		"created_at": user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile retrieved successfully",
		"user":    response,
	})
}

// Logout удаляет токен из куки
func (ctrl *UserController) Logout(c *gin.Context) {
	// Удаляем куки
	c.SetCookie(
		"auth_token", // name
		"",           // value (пустое)
		-1,           // maxAge (удалить)
		"/",          // path
		"",           // domain
		true,         // secure
		true,         // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (ctrl *UserController) LogIn(c *gin.Context) {
	// Get the email and pass off req body
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ctrl.userService.LogIn(body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

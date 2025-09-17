package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {

	// Get the email or password from req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		// Set error ansver
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
	}

	// Create the user
	user := models.User{Email: body.Email, Password: string(hash)}

	resu

}

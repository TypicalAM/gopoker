package routes

import (
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterData is the data that is sent to the register route
type RegisterData struct {
	Username string
	Password string
}

// Register allows the user to register a new account
func (controller *Controller) Register(c *gin.Context) {
	var data RegisterData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if data.Username == "" || data.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if len(data.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

	user := models.User{Username: data.Username}
	res := controller.db.Where(&user).First(&user)
	if res.Error == nil || res.RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	// TODO: The GameID hack is a temporary fix
	user.Password = string(hashedPassword)
	user.GameID = 1

	res = controller.db.Save(&user)
	if res.Error != nil || res.RowsAffected == 0 {
		log.Println(res.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Success"})
}

package routes

import (
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	passwordErr     = "Your password must be at least 8 characters long or longer"
	registerErr     = "There was an error registering your account. Please try again later."
	registerSuccess = "Your account has been created. You may now log in."
)

// Register renders the HTML content of the register page.
func (controller Controller) Register(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Register"
	c.HTML(http.StatusOK, "register.html", pd)
}

// RegisterPost handles requests to register a new user.
func (controller Controller) RegisterPost(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Register"

	// Check if the password is at least 8 characters long
	password := c.PostForm("password")
	if len(password) < 8 {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: passwordErr,
		})
		c.HTML(http.StatusBadRequest, "register.html", pd)
		return
	}

	// Let's hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: registerErr,
		})
		c.HTML(http.StatusInternalServerError, "register.html", pd)
		return
	}

	// Create the user
	username := c.PostForm("username")
	user := models.User{
		Username: username,
		// TODO: REMOVE THIS
		GameID:   1,
	}

	res := controller.db.Where(&user).First(&user)
	if (res.Error != nil) && (res.Error != gorm.ErrRecordNotFound || res.RowsAffected > 0) {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: registerErr,
		})
		log.Println(res.Error)
		c.HTML(http.StatusInternalServerError, "register.html", pd)
		return
	}

	user.Password = string(hashedPassword)
	res = controller.db.Save(&user)
	if res.Error != nil || res.RowsAffected == 0 {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: registerErr,
		})
		log.Println(res.Error)
		c.HTML(http.StatusInternalServerError, "register.html", pd)
		return
	}

	log.Println("User created:", user.Username)
	pd.Messages = append(pd.Messages, Message{
		Type:    "success",
		Content: registerSuccess,
	})

	c.HTML(http.StatusOK, "register.html", pd)
}

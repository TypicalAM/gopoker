package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/ulid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	loginError = "Could nt log you in. Please double check your username and password."
)

// Login renders the HTML content of the login page.
func (controller Controller) Login(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Login"
	c.HTML(http.StatusOK, "login.html", pd)
}

// LoginPost handles requests to log in a user.
func (controller Controller) LoginPost(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Login"

	username := c.PostForm("username")
	user := models.User{Username: username}

	res := controller.db.Where(&user).First(&user)
	if res.Error != nil || res.RowsAffected == 0 {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: loginError,
		})
		c.HTML(http.StatusBadRequest, "login.html", pd)
		return
	}

	password := c.PostForm("password")
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: loginError,
		})
		c.HTML(http.StatusBadRequest, "login.html", pd)
		return
	}

	sessionID := ulid.Generate()
	ses := models.Session{
		Identifier: sessionID,
		UserID:     user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	res = controller.db.Create(&ses)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: loginError,
		})
		c.HTML(http.StatusBadRequest, "login.html", pd)
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.SessionIDKey, sessionID)

	err = session.Save()
	if err != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: loginError,
		})
		c.HTML(http.StatusBadRequest, "login.html", pd)
		return
	}

	log.Println("User logged in:", user.Username)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

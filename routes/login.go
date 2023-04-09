package routes

import (
	"net/http"
	"time"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/ulid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginData is the data that is sent to the login route
type LoginData struct {
	Username string
	Password string
}

// Login is the route that handles the login
func (controller Controller) Login(c *gin.Context) {
	var data LoginData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if data.Username == "" || data.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	user := models.User{Username: data.Username}
	res := controller.db.Where(&user).First(&user)
	if res.Error != nil || res.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	sessionID := ulid.Generate()
	ses := models.Session{
		Identifier: sessionID,
		UserID:     user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	res = controller.db.Create(&ses)
	if res.Error != nil || res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.SessionIDKey, sessionID)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

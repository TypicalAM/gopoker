package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var userNotFoundErr = fmt.Errorf("User not found")

func (controller Controller) getUser(c *gin.Context) (*models.User, error) {
	var user models.User

	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		return nil, userNotFoundErr
	}

	if userID != nil {
		user.ID = userID.(uint)
		res := controller.db.First(&user)
		if res.Error == nil {
			return &user, nil
		}
	}

	return nil, userNotFoundErr
}

// Lobby is the lobby page
func (controller Controller) Lobby(c *gin.Context) {
	pd := controller.DefaultPageData(c)

	user, _ := controller.getUser(c)
	log.Println(user.Username)
	session := sessions.Default(c)
	gameIDInterface := session.Get(models.GameIDKey)
	if gameID, ok := gameIDInterface.(string); ok {
		log.Println("GameID found, redirecting to game: ", gameID)
		c.Redirect(http.StatusFound, "/game/id/"+gameID)
		return
	}

	c.HTML(http.StatusOK, "lobby.html", pd)
}

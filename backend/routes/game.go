package routes

import (
	"errors"
	"net/http"

	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var IncorrectGameErr = errors.New("incorrect game")

// Game is the websocket game connection
func (controller Controller) Game(c *gin.Context) {
	session := sessions.Default(c)

	// Check if the user is in the game
	gameModel, user, err := ensureCorrectGame(controller.db, session, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "incorrect game",
		})
		return
	}

	game.ServeWs(controller.hub, controller.db, c, gameModel, user)
}

// ensureCorrectGame checks if the user is in the game and the game exists
func ensureCorrectGame(db *gorm.DB, session sessions.Session, c *gin.Context) (*models.Game, *models.User, error) {
	gameID := c.Param("id")

	gameIDInterface := session.Get(models.GameIDKey)
	if gameIDInterface == nil || gameIDInterface.(string) != gameID {
		return nil, nil, IncorrectGameErr
	}

	var game models.Game
	res := db.Model(&models.Game{}).Preload("Players").Where("uuid = ?", gameID).Find(&game)
	if res.Error != nil {
		return nil, nil, IncorrectGameErr
	}

	var user models.User
	res = db.Model(&models.User{}).Where("id = ?", c.MustGet(middleware.UserIDKey)).Find(&user)
	if res.Error != nil {
		return nil, nil, IncorrectGameErr
	}

	var found bool
	for _, player := range game.Players {
		if player.ID == user.ID {
			found = true
			break
		}
	}

	if !found {
		return nil, nil, IncorrectGameErr
	}

	return &game, &user, nil
}

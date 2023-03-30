package routes

import (
	"errors"
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var IncorrectGameErr = errors.New("Incorrect game")

// ensureCorrectGame checks if the user is in the game and the game exists
func ensureCorrectGame(db *gorm.DB, session sessions.Session, c *gin.Context, pd *PageData) (*models.Game, *models.User, error) {
	gameID := c.Param("id")

	// Check if the GameID matches the one in the session
	gameIDInterface := session.Get(models.GameIDKey)
	if gameIDInterface == nil || gameIDInterface.(string) != gameID {
		return nil, nil, IncorrectGameErr
	}

	// Check if this game exists
	var game models.Game
	if res := db.Model(&models.Game{}).Preload("Players").Where("uuid = ?", gameID).Find(&game); res.Error != nil {
		return nil, nil, IncorrectGameErr
	}

	// Check if the user has its key
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		return nil, nil, IncorrectGameErr
	}

	// Check if the user exists
	var user models.User
	user.ID = userID.(uint)
	if res := db.First(&user); res.Error != nil {
		return nil, nil, IncorrectGameErr
	}

	// Check if the user is in the game
	var found bool
	for _, player := range game.Players {
		if player.ID == user.ID {
			found = true
			break
		}
	}

	// If the user is not in the game, redirect to the lobby
	if !found {
		return nil, nil, IncorrectGameErr
	}

	return &game, &user, nil
}

// Add a message and redirect to the lobby
func redirectToLobby(session sessions.Session, c *gin.Context, pd *PageData) {
	log.Println("Redirecting from game to lobby.")
	session.Delete(models.GameIDKey)
	session.Save()
	c.Redirect(http.StatusFound, "/game/lobby")
	return
}

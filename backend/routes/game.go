package routes

import (
	"errors"
	"net/http"

	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var incorrectGameErr = errors.New("incorrect game")

// Game is the websocket game connection
func (con controller) Game(c *gin.Context) {
	user, err := con.getUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	gameModel, err := ensureCorrectGame(con.db, user, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect game"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't upgrade to websocket"})
		return
	}

	game.Connect(con.hub, con.db, conn, gameModel, user)
}

// ensureCorrectGame checks if the user is in the game and the game exists
func ensureCorrectGame(db *gorm.DB, user *models.User, c *gin.Context) (*models.Game, error) {
	session := sessions.Default(c)
	gameID := c.Param("id")

	gameIDInterface := session.Get(models.GameIDKey)
	if gameIDInterface == nil || gameIDInterface.(string) != gameID {
		return nil, incorrectGameErr
	}

	var game models.Game
	res := db.Model(&models.Game{}).Preload("Players").Where("uuid = ?", gameID).Find(&game)
	if res.Error != nil {
		return nil, incorrectGameErr
	}

	for _, player := range game.Players {
		if player.ID == user.ID {
			return &game, nil
		}
	}

	return nil, incorrectGameErr
}

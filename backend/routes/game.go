package routes

import (
	"errors"
	"log"
	"net/http"

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

	game, err := con.ensureCorrectGame(con.db, user, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect game"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't upgrade to websocket"})
		return
	}

	con.gameSrv.Connect(conn, game, user)
}

// ensureCorrectGame checks if the user is in the game and the game exists
func (con controller) ensureCorrectGame(db *gorm.DB, user *models.User, c *gin.Context) (*models.Game, error) {
	log.Println("Adding a player because of the link")
	session := sessions.Default(c)
	gameID := c.Param("id")

	var game models.Game
	res := db.Model(&models.Game{}).Preload("Players").Where("uuid = ?", gameID).Find(&game)
	if res.Error != nil {
		return nil, incorrectGameErr
	}

	gameIDInterface := session.Get(models.GameIDKey)
	if gameIDInterface == nil || gameIDInterface.(string) != gameID {
		// Can we add a player?
		if game.Playing || len(game.Players) == con.config.GamePlayerCap {
			return nil, incorrectGameErr
		}

		game.Players = append(game.Players, *user)
		res = con.db.Save(&game)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "There was an error adding you to the game. Please try again later.",
			})
			return nil, res.Error
		}

		session.Set(models.GameIDKey, game.UUID)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "There was an error saving your session. Please try again later.",
			})
			return nil, err
		}

		return &game, nil
	}

	for _, player := range game.Players {
		if player.ID == user.ID {
			return &game, nil
		}
	}

	return nil, incorrectGameErr
}

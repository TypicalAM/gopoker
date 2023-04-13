package routes

import (
	"net/http"
	"sort"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Queue allows the user to join a game queue
func (controller Controller) Queue(c *gin.Context) {
	var games []models.Game
	res := controller.db.Model(&models.Game{}).Preload("Players").Where("playing = ?", false).Find(&games)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "There was an error finding games. Please try again later.",
		})
		return
	}

	var user models.User
	res = controller.db.Where("id = ?", c.MustGet(middleware.UserIDKey)).First(&user)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "There was an error finding your user. Please try again later.",
		})
	}

	session := sessions.Default(c)
	gameIDInterface := session.Get(models.GameIDKey)
	if gameID, ok := gameIDInterface.(string); ok {
		var game models.Game
		res = controller.db.Where("uuid = ?", gameID).First(&game)
		if res.Error != nil {
			session.Set(models.GameIDKey, nil)
		} else {
			userInGame := false
			for _, player := range game.Players {
				if player.ID == user.ID {
					userInGame = true
				}
			}

			if userInGame {
				c.JSON(http.StatusOK, gin.H{
					"message": "You are already in a game.",
					"uuid":    game.UUID,
				})
				return
			}

			session.Delete(models.GameIDKey)
		}
	}

	if len(games) == 0 {
		controller.createNewGame(c, &user)
		return
	}

	sort.Slice(games, func(i, j int) bool {
		return len(games[i].Players) < len(games[j].Players)
	})

	for i, game := range games {
		// Check if we didn't fully fill the game in the meantime
		if game.Playing || len(game.Players) == controller.config.GamePlayerCap {
			continue
		}

		// Add the user to the game
		games[i].Players = append(games[i].Players, user)
		res = controller.db.Save(&games[i])
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "There was an error adding you to the game. Please try again later.",
			})
			return
		}

		session := sessions.Default(c)
		session.Set(models.GameIDKey, game.UUID)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "There was an error saving your session. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uuid": game.UUID,
		})

		return
	}
}

// createNewGame creates a new game and adds the user to it
func (controller Controller) createNewGame(c *gin.Context, user *models.User) {
	newGameUUID := uuid.New().String()
	game := models.Game{
		Playing: false,
		UUID:    newGameUUID,
		Players: []models.User{*user},
	}

	res := controller.db.Model(&models.Game{}).Preload("Players").Create(&game)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "There was an error creating a new game. Please try again later.",
		})
		return
	}

	session := sessions.Default(c)
	session.Set(models.GameIDKey, newGameUUID)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "There was an error saving your session. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uuid": game.UUID,
	})
}

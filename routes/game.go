package routes

import (
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Game is the route used for the game page
func (controller Controller) Game(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Game"
	gameID := c.Param("id")

	log.Println("GameID: ", gameID)

	// Check if the GameID matches the one in the session
	session := sessions.Default(c)
	gameIDInterface := session.Get(models.GameIDKey)
	if gameIDInterface == nil || gameIDInterface.(string) != gameID {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "You are not in this game.",
		})
		log.Println("Redirecting to lobby: 1")
		session.Delete(models.GameIDKey)
		session.Save()
		c.Redirect(http.StatusFound, "/game/lobby")
		return
	}

	// Check if this game exists
	var game models.Game
	res := controller.db.Model(&models.Game{}).Preload("Players").Where("uuid = ?", gameID).Find(&game)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "This game does not exist.",
		})
		session.Delete(models.GameIDKey)
		session.Save()
		log.Println("Redirecting to lobby: 2")
		c.Redirect(http.StatusFound, "/game/lobby")
		return
	}

	// Check if the user has its key
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "You are not logged in.",
		})

		session.Delete(models.GameIDKey)
		session.Save()
		log.Println("Redirecting to lobby: 3")
		c.Redirect(http.StatusFound, "/game/lobby")
		return
	}

	// Check if the user exists
	var user models.User
	user.ID = userID.(uint)
	res = controller.db.First(&user)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "You are not logged in.",
		})

		session.Delete(models.GameIDKey)
		session.Save()
		log.Println("Redirecting to lobby: 4")
		c.Redirect(http.StatusFound, "/game/lobby")
		return
	}

	// Check if the user is in the game
	var found bool
	log.Println("Amount of players: ", len(game.Players))
	for _, player := range game.Players {
		if player.ID == user.ID {
			found = true
			break
		}
	}

	if !found {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "You are not in this game.",
		})

		session.Delete(models.GameIDKey)
		session.Save()
		log.Println("Redirecting to lobby: 5")
		c.Redirect(http.StatusFound, "/game/lobby")
		return
	}

	// Show the game to the players
	for _, player := range game.Players {
		pd.Messages = append(pd.Messages, Message{
			Type:    "success",
			Content: player.Username + " is in the game!",
		})
	}

	log.Println("A player is shown the game screen with gameID: ", gameID)
	c.HTML(http.StatusOK, "game.html", pd)
}

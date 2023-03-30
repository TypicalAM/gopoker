package routes

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Game is the route used for the game page
func (controller Controller) Game(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Game"
	session := sessions.Default(c)

	// Check if the user is in the game
	game, user, err := ensureCorrectGame(controller.db, session, c, &pd)
	if err != nil {
		redirectToLobby(session, c, &pd)
		return
	}

	// Show the game to the players
	for _, player := range game.Players {
		if player.ID == user.ID {
			pd.Messages = append(pd.Messages, Message{
				Type:    "success",
				Content: "You are in the game!",
			})
		} else {
			pd.Messages = append(pd.Messages, Message{
				Type:    "success",
				Content: player.Username + " is in the game!",
			})
		}
	}

	log.Println("A player is shown the game screen with gameID: ", game.UUID)
	c.HTML(http.StatusOK, "game.html", pd)
}

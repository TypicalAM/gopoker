package routes

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LeaveGame is the route used for leaving a game
func (controller Controller) LeaveGame(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	pd.Title = "Leave Game"
	session := sessions.Default(c)

	// Check if the user is in the game
	game, user, err := ensureCorrectGame(controller.db, session, c, &pd)
	if err != nil {
		redirectToLobby(session, c, &pd)
		return
	}

	// Remove the game if there are no players left
	log.Println("Player", user.Username, "is leaving the game with gameID:", game.UUID)
	if len(game.Players) == 1 {
		log.Println("The game is being deleted because there are no players left")
		controller.db.Model(&game).Association("Players").Clear()
		controller.db.Delete(&game)
		redirectToLobby(session, c, &pd)
		return
	}

	// Remove the player from the game
	for i, player := range game.Players {
		if player.ID == user.ID {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			break
		}
	}

	// Save the game
	controller.db.Model(&game).Association("Players").Replace(game.Players)
	redirectToLobby(session, c, &pd)
}

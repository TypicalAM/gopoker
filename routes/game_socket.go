package routes

import (
	"github.com/TypicalAM/gopoker/websockets"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// GameSocket handles the websocket connection for the game
func (controller Controller) GameSocket(c *gin.Context) {
	pd := controller.DefaultPageData(c)
	session := sessions.Default(c)

	// Check if the user is in the game
	game, user, err := ensureCorrectGame(controller.db, session, c, &pd)
	if err != nil {
		redirectToLobby(session, c, &pd)
		return
	}

	// Serve the websocket
	websockets.ServeWs(controller.hub, controller.db, c, game, user)
}

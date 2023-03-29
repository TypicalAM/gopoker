package routes

import (
	"log"
	"net/http"
	"sort"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Queue is the queue page, users are redirected here when they want to play a game.
// It searches for a game to join, and if none are found, it creates a new game.
func (controller Controller) Queue(c *gin.Context) {
	// Get the default page data
	pd := controller.DefaultPageData(c)

	// Get all the games which have playing = false
	// If there are no games, create a new game
	games := []models.Game{}
	res := controller.db.Where("playing = ?", false).Find(&games)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "There was an error finding games. Please try again later.",
		})
		c.HTML(http.StatusInternalServerError, "queue.html", pd)
		return
	}

	// Get the user from the request
	user := models.User{}
	res = controller.db.Where("id = ?", c.MustGet(middleware.UserIDKey)).First(&user)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "There was an error finding your user. Please try again later.",
		})
		c.HTML(http.StatusInternalServerError, "queue.html", pd)
		return
	}

	log.Println("Found games: ", len(games))
	if len(games) == 0 {
		controller.createNewGame(c, &user)
		return
	} else {
		// Sort the games by the number of players descending
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
				pd.Messages = append(pd.Messages, Message{
					Type:    "error",
					Content: "There was an error joining the game. Please try again later.",
				})
				c.HTML(http.StatusInternalServerError, "queue.html", pd)
				return
			}

			// Set the GameID of the user to represent that they are in a game
			session := sessions.Default(c)
			session.Set(models.GameIDKey, game.UUID)
			session.Save()

			// Redirect to the game
			log.Println("Found, redirecting to game: ", game.UUID)
			c.Redirect(http.StatusFound, "/game/id/"+game.UUID)
			return
		}
	}

	c.HTML(http.StatusOK, "queue.html", pd)
}

func (controller *Controller) createNewGame(c *gin.Context, user *models.User) {
	pd := controller.DefaultPageData(c)
	session := sessions.Default(c)

	// Create a new game
	newGameUUID := uuid.New().String()
	game := models.Game{
		Playing: false,
		UUID:    newGameUUID,
		Players: []models.User{*user},
	}

	// Try to save the new game
	res := controller.db.Create(&game)
	if res.Error != nil {
		pd.Messages = append(pd.Messages, Message{
			Type:    "error",
			Content: "There was an error creating a new game. Please try again later.",
		})
		c.HTML(http.StatusInternalServerError, "queue.html", pd)
		return
	}

	// Set the GameID of the user to represent that they are in a game
	session.Set(models.GameIDKey, newGameUUID)
	session.Save()

	// Redirect to the game page
	log.Println("Created, redirecting to game: ", newGameUUID)
	c.Redirect(http.StatusFound, "/game/id/"+newGameUUID)
}

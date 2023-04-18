package game

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/TypicalAM/gopoker/models"
)

// Game represents an instance of a game server.
type Game struct {
	UUID    string
	hub     *Hub
	texas   *TexasHoldEm
	clients []*Client
}

// newGame creates a new game.
func newGame(hub *Hub, uuid string) *Game {
	return &Game{
		UUID:  uuid,
		hub:   hub,
		texas: NewTexasHoldEm(),
	}
}

// addClient adds a client to the game.
func (g *Game) addClient(c *Client) {
	log.Printf("Adding client %s to game %s", c.userModel.Username, g.UUID)
	g.clients = append(g.clients, c)

	// Let's try adding the client to the game
	err := g.texas.AddPlayer(c.userModel.Username, 100)
	if err != nil && errors.Is(err, InternalErr) {
		log.Println("Adding a client to a game in which the player already exists")
	}

	// Let's try to start the game
	if err = g.texas.StartGame(); err != nil {
		log.Printf("Error starting game: %s", err)
		return
	}

	// Update the game in the database
	if res := c.db.Model(&models.Game{}).Where("uuid = ?", c.game.UUID).Update("Playing", true); res.Error != nil {
		log.Printf("Error updating game: %s", res.Error)
		return
	}

	// Broadcast the game state
	log.Println("The game has been started, broadcasting the state")
	g.broadcastState()
}

// handleMessage handles a message from a client.
func (g *Game) handleMessage(client *Client, gameMsg *GameMessage) {
	switch gameMsg.Type {
	case MsgAction:
		var action pokerAction
		if val, ok := actionMap[gameMsg.Data]; !ok {
			g.sendMessageToClient(client, &GameMessage{
				Type: MsgError,
				Data: "Invalid action",
			})
			return
		} else {
			action = val
		}

		if err := g.texas.AdvanceState(client.userModel.Username, action); err != nil {
			g.sendMessageToClient(client, &GameMessage{
				Type: MsgError,
				Data: err.Error(),
			})
		}

		g.broadcastState()

	default:
		g.sendMessageToClient(client, &GameMessage{
			Type: MsgError,
			Data: "Incorrect action",
		})
	}
}

// sendMessageToClient sends a message to a client.
func (g *Game) sendMessageToClient(client *Client, gameMsg *GameMessage) {
	select {
	case client.send <- *gameMsg:
	default:
		g.removeClient(client)
	}
}

// broadcastState sends a state message to every client
func (g *Game) broadcastState() {
	for _, client := range g.clients {
		sanitizedState := g.texas.SanitizeState(client.userModel.Username)
		stateBytes, _ := json.Marshal(sanitizedState)
		g.sendMessageToClient(client, &GameMessage{
			Type: MsgState,
			Data: string(stateBytes),
		})
	}
}

// removeClient removes a client from the game.
func (g *Game) removeClient(c *Client) error {
	close(c.send)
	for i, client := range g.clients {
		if client == c {
			g.clients = append(g.clients[:i], g.clients[i+1:]...)
			break
		}
	}

	err := g.texas.Disconnect(c.userModel.Username)
	if err != nil {
		if errors.Is(err, PlayerNotInGameErr) {
			log.Println("Removing a client from a game in which the player does not exist")
		} else if errors.Is(err, OwnTurnDisconnectErr) {
			log.Println("Disconnecting a client in which it is their turn, advancing the game")
			g.broadcastState()
		} else {
			return err
		}
	}

	if g.texas.ShouldBeDisbanded() {
		log.Println("The game should be disbanded, removing it from the hub")
		g.hub.deleteGame(g.UUID)
	}

	return nil
}

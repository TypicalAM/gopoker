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
	err = g.texas.StartGame()
	if err != nil && errors.Is(err, NotEnoughPlayersErr) {
		log.Println("There are not enough players to start the game")
		return
	}

	// Update the game in the database
	if res := c.db.Model(&models.Game{}).Where("uuid = ?", c.game.UUID).Update("Playing", true); res.Error != nil {
		log.Printf("Error updating game: %s", res.Error)
	}

	// Broadcast the game state
	log.Println("The game has been started, broadcasting the state")
	g.broadcastState()
}

// handleMessage handles a message from a client.
func (g *Game) handleMessage(client *Client, gameMsg *GameMessage) {
	switch gameMsg.Type {
	case MsgInput:
		// TODO
		panic("msgInput not implemented")

	case MsgAction:
		// Convert string to pokeraction
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

	default:
		g.sendMessageToClient(client, &GameMessage{
			Type: MsgError,
			Data: "Incorrect action",
		})
		return
	}

	// Broadcast the game state
	g.broadcastState()
}

// sendMessageToClient sends a message to a client.
func (g *Game) sendMessageToClient(client *Client, gameMsg *GameMessage) {
	select {
	case client.send <- *gameMsg:
	default:
		// TODO: Just disconnect the client?
		close(client.send)
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
func (g *Game) removeClient(c *Client) {
	for i, client := range g.clients {
		if client == c {
			g.clients = append(g.clients[:i], g.clients[i+1:]...)
			break
		}
	}

	err := g.texas.Disconnect(c.userModel.Username)
	if err != nil && errors.Is(err, PlayerNotInGameErr) {
		log.Println("Removing a client from a game in which the player does not exist")
	}
}

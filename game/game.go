package game

import (
	"encoding/json"
	"errors"
	"log"
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
		texas: newTexasHoldEm(),
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

	// Broadcast the game state
	g.broadcastState()
}

// handleMessage handles a message from a client.
func (g *Game) handleMessage(client *Client, gameMsg *GameMessage) {
	switch gameMsg.Type {
	case msgInput:
		// TODO
		panic("msgInput not implemented")

	case msgAction:
		// Convert string to pokeraction
		var action pokerAction
		if val, ok := interface{}(gameMsg.Data).(pokerAction); !ok {
			g.sendMessageToClient(client, &GameMessage{
				Type: msgError,
				Data: "Invalid action",
			})
			return
		} else {
			action = val
		}

		if err := g.texas.AdvanceState(client.userModel.Username, action); err != nil {
			g.sendMessageToClient(client, &GameMessage{
				Type: msgError,
				Data: err.Error(),
			})
		}

	default:
		g.sendMessageToClient(client, &GameMessage{
			Type: msgError,
			Data: "Incorrect action",
		})
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
			Type: msgState,
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

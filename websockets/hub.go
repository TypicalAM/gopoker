package websockets

import (
	"fmt"
	"log"
	"time"
)

// statusInterval is the interval at which the hub will send status updates.
const statusInterval = 10 * time.Second

// thankYouMessage is the message that will be sent to clients.
var thankYouMsg = GameMessage{
	Type: msgStatus,
	Data: "Thank you for playing!",
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	games      map[string]*Game
	broadcast  chan GameMessage
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan GameMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		games:      make(map[string]*Game),
	}
}

// Run starts the hub.
func (h *Hub) Run() {
	ticker := time.NewTicker(statusInterval)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.sendThankYouMessages()
		}
	}
}

// registerClient adds the client to the hub.
func (h *Hub) registerClient(client *Client) {
	// If the game already exists, we add the client to the game.
	if game, ok := h.games[client.game.UUID]; ok {
		log.Println(fmt.Sprintf("[%s] Adding player %s to game", client.game.UUID, client.player.Username))
		game.Players = append(game.Players, client)
		h.games[client.game.UUID] = game
		return
	}

	// If the game doesn't exist, we create a new one.
	log.Println(fmt.Sprintf("[%s] Creating new game and adding player %s", client.game.UUID, client.player.Username))
	h.games[client.game.UUID] = newGame(client.game.UUID, client)
}

// unregisterClient removes the client from the hub and closes the game if needed.
func (h *Hub) unregisterClient(client *Client) {
	// If the game doesn't exist, we do nothing.
	if _, ok := h.games[client.game.UUID]; !ok {
		return
	}

	// We have to close the client's send channel.
	close(client.send)

	// If the game exists, we remove the client from the game.
	game := h.games[client.game.UUID]
	for i, c := range game.Players {
		if c == client {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			break
		}
	}

	// If the game is empty, we close the game.
	if len(game.Players) == 0 {
		delete(h.games, client.game.UUID)
		return
	}

	// If the game is not empty, we update the game.
	h.games[client.game.UUID] = game
}

// broadcastMessage sends the message to all clients.
func (h *Hub) broadcastMessage(message GameMessage) {
	for _, game := range h.games {
		for _, client := range game.Players {
			select {
			case client.send <- message:
			default:
				h.unregisterClient(client)
			}
		}
	}
}

// sendThankYouMessages sends a thank you message to all clients.
func (h *Hub) sendThankYouMessages() {
	for _, game := range h.games {
		for _, client := range game.Players {
			select {
			case client.send <- thankYouMsg:
			default:
				h.unregisterClient(client)
			}
		}
	}
}

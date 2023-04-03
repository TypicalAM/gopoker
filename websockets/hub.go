package websockets

import (
	"log"
	"time"
)

var (
	// thankYouMsg is the message sent to the client every thankYouMsgInterval.
	thankYouMsg = GameMessage{
		Type: msgStatus,
		Data: "Thank you for playing!",
	}

	// thankYouMsgInterval is the interval at which the thank you message is sent.
	thankYouMsgInterval = 50 * time.Second
)

// Hub maintains the set of active clients and game servers, it also broadcasts messages to the clients.
type Hub struct {
	games      gameStore
	broadcast  chan GameMessageWithSender
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new hub.
func NewHub() *Hub {
	return &Hub{
		games:      newGameStore(),
		broadcast:  make(chan GameMessageWithSender, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub.
func (h *Hub) Run() {
	ticker := time.NewTicker(thankYouMsgInterval)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case gameMsg := <-h.broadcast:
			h.handleMessage(gameMsg.Sender, &gameMsg.Message)

		case <-ticker.C:
			h.sendThankYouMsg()
		}
	}
}

// registerClient registers a client and creates a new game if it doesn't exist.
func (h *Hub) registerClient(client *Client) {
	game, ok := h.games.load(client.game.UUID)
	if !ok {
		game = newGame(h, client.game.UUID)
		go game.run()
	}

	game.addPlayer(client)
	h.games.save(game)
}

// unregisterClient unregisters a client and removes the game if it doesn't have any clients.
func (h *Hub) unregisterClient(client *Client) {
	game, ok := h.games.load(client.game.UUID)
	if !ok {
		return
	}

	game.removePlayer(client)
	if len(game.Players) == 0 {
		h.games.delete(client.game.UUID)
	}
}

// handleMessage handles a game message sent by a client.
func (h *Hub) handleMessage(client *Client, gameMsg *GameMessage) {
	game, ok := h.games.load(client.game.UUID)
	if !ok {
		return
	}

	game.handleMessage(client, gameMsg)
}

// sendThankYouMsg sends a thank you message to all the clients.
func (h *Hub) sendThankYouMsg() {
	log.Println("Sending thank you message to all the clients")
	for _, game := range h.games.loadAll() {
		for _, client := range game.Players {
			select {
			case client.send <- thankYouMsg:
			default:
				h.unregisterClient(client)
			}
		}
	}
}

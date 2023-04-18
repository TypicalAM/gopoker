package game

import "log"

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
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case gameMsg := <-h.broadcast:
			h.handleMessage(gameMsg.Sender, &gameMsg.Message)
		}
	}
}

// registerClient registers a client and creates a new game if it doesn't exist.
func (h *Hub) registerClient(client *Client) {
	game, ok := h.games.load(client.game.UUID)
	if !ok {
		game = newGame(h, client.game.UUID)
	}

	game.addClient(client)
	h.games.save(game)
}

// unregisterClient unregisters a client and removes the game if it doesn't have any clients.
func (h *Hub) unregisterClient(client *Client) {
	game, ok := h.games.load(client.game.UUID)
	if !ok {
		return
	}

	if err := game.removeClient(client); err != nil {
		log.Printf("error removing client from game: %v", err)
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

// deleteGame deletes a game.
func (h *Hub) deleteGame(uuid string) {
	h.games.delete(uuid)
}

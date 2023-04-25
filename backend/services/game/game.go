package game

import (
	"log"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Server maintains the set of active clients and lobbies.
type Server struct {
	db              *gorm.DB
	games           gameStore
	registerQueue   chan *Client
	unregisterQueue chan *Client
}

// New creates a new game server.
func New(db *gorm.DB) *Server {
	return &Server{
		db:              db,
		games:           newGameStore(),
		registerQueue:   make(chan *Client),
		unregisterQueue: make(chan *Client),
	}
}

// Run starts the game server.
func (srv *Server) Run() {
	for {
		select {
		case client := <-srv.registerQueue:
			srv.register(client)

		case client := <-srv.unregisterQueue:
			srv.unregisterClient(client)
		}
	}
}

// Connect creates a new client and enqueues it for registration.
func (srv *Server) Connect(conn *websocket.Conn, game *models.Game, user *models.User) {
	lobby, ok := srv.games.load(game.UUID)
	if !ok {
		lobby = newLobby(srv, game.UUID)
		srv.games.save(lobby)
	}

	client := newClient(srv, lobby, conn, user)
	srv.registerQueue <- client
	go client.writeLoop()
	go client.readLoop()
}

// register registers a client with the game server and assigns it to a lobby.
func (srv *Server) register(client *Client) {
	log.Printf("[%s] Registering client", client.user.Username)
	client.lobby.addClient(client)
}

// unregister removes a client from the game server and deletes a
// lobby if it is empty.
func (srv *Server) unregisterClient(client *Client) {
	log.Printf("[%s] Unregistering client", client.user.Username)
	client.conn.Close()
	close(client.send)

	game, ok := srv.games.load(client.lobby.uuid)
	if !ok {
		// The game has already been deleted
		return
	}

	if err := game.disconnect(client); err != nil {
		log.Printf("[%s] Error disconnecting client: %s", client.lobby.uuid[:10], err)
		return
	}

	if game.isEmpty() {
		log.Printf("[%s] Deleting game", client.lobby.uuid[:10])
		srv.deleteGame(client.lobby.uuid)
	}
}

// startGame starts a game.
func (srv *Server) startGame(uuid string) {
	if res := srv.db.Model(&models.Game{}).Where("uuid = ?", uuid).Update("Playing", true); res.Error != nil {
		log.Printf("[%s] Error updating game: %s", uuid[:10], res.Error)
		return
	}
}

// deleteGame deletes a game.
func (srv *Server) deleteGame(uuid string) {
	srv.games.delete(uuid)
	if res := srv.db.Delete(&models.Game{}, "uuid = ?", uuid); res.Error != nil {
		log.Printf("[%s] Error deleting game model: %s", uuid[:10], res.Error)
	}
	log.Printf("[%s] Ended & Deleted", uuid[:10])
}

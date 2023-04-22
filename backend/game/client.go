package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	//pongWait = 60 * time.Second
	pongWait = 2 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub       *Hub
	db        *gorm.DB
	userModel *models.User
	game      *models.Game
	conn      *websocket.Conn
	send      chan GameMessage
}

// Connect takes the websocket connection and bootstraps the client
func Connect(hub *Hub, db *gorm.DB, conn *websocket.Conn, game *models.Game, user *models.User) {
	client := &Client{
		hub:       hub,
		db:        db,
		conn:      conn,
		userModel: user,
		game:      game,
		send:      make(chan GameMessage, 256),
	}

	client.hub.register <- client
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Message handling
		var gameMsg *GameMessage
		if err := json.Unmarshal(message, &gameMsg); err != nil {
			log.Println(fmt.Sprintf("[%s] Invalid message from %s", c.game.UUID, c.userModel.Username))
			log.Println(fmt.Sprintf("The following error occurred: %s", err))
			continue
		}

		c.hub.broadcast <- GameMessageWithSender{
			Message: *gameMsg,
			Sender:  c,
		}
	}
}

// msgType is the type of the game message.
type msgType string

const (
	MsgState  msgType = "state"
	MsgError          = "error"
	MsgInput          = "input"
	MsgAction         = "action"
)

// GameMessage is a message that is used to communicate between the player and the game server.
type GameMessage struct {
	Type msgType `json:"type"`
	Data string  `json:"data"`
}

// GameMessageWithSender is a message that is sent to the hub to be broadcasted.
type GameMessageWithSender struct {
	Message GameMessage
	Sender  *Client
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Println("Couldn't marshal the message:", err)
				return
			}

			w.Write(messageBytes)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				messageBytes, err := json.Marshal(<-c.send)
				if err != nil {
					log.Println("Couldn't marshal the message:", err)
					return
				}

				w.Write(messageBytes)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}



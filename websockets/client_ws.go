package websockets

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-gonic/gin"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TODO: Check the origin of the request
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
	db  *gorm.DB

	// player is the player that this client is representing
	player *models.User

	// game is the game that this client is in
	game *models.Game

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
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
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Printf("Received message from %s: %s", c.player.Username, message)
		if c.parseMessage(string(message)) {
			continue
		}

		c.hub.broadcast <- message
	}
}

// parseMessage parses a message from the client and returns true if the message is a special one
func (c *Client) parseMessage(message string) bool {
	if len(message) < 7 {
		return false
	}

	switch message[:6] {
	case "uinput":
		log.Printf("Received credit card number from %s: %s", c.player.Username, message[7:])
		c.player.UnsecuredCreditcard = string(message[7:])
		c.db.Save(c.player)
		return true

	case "Punch ":
		log.Printf("Received punch from %s: %s", c.player.Username, message[6:])
		c.hub.broadcast <- []byte("status:" + c.player.Username + " punched " + message[6:])
		return true
	}

	return false
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
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
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			b := strings.Builder{}
			b.WriteString("action:")
			for i, player := range c.game.Players {
				if player.ID == c.player.ID {
					continue
				}

				b.WriteString("Punch ")
				b.WriteString(player.Username)
				if i < len(c.game.Players)-1 {
					b.WriteString(",")
				}
			}

			c.conn.WriteMessage(websocket.TextMessage, []byte(b.String()))

			// Unsecure credit card sniff
			b.Reset()
			if strings.TrimSpace(c.player.UnsecuredCreditcard) == "" {
				b.WriteString("uinput:Credit card number")
				c.conn.WriteMessage(websocket.TextMessage, []byte(b.String()))
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, db *gorm.DB, c *gin.Context, game *models.Game, user *models.User) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}

	log.Println("New websocket connection, creating a client for", user.Username)

	client := &Client{
		hub:    hub,
		db:     db,
		player: user,
		game:   game,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/TypicalAM/gopoker/models"
	"github.com/gorilla/websocket"
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

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	srv   *Server
	lobby *lobby
	user  *models.User
	conn  *websocket.Conn
	send  chan GameMessage
}

// Connect takes the websocket connection and bootstraps the client
func newClient(srv *Server, l *lobby, conn *websocket.Conn, user *models.User) *Client {
	return &Client{
		srv:  srv,
		conn: conn,
		lobby: l,
		user: user,
		send: make(chan GameMessage, 256),
	}
}

// readLoop pumps messages from the websocket connection to the hub.
func (c *Client) readLoop() {
	defer func() {
		c.srv.unregisterQueue <- c
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
			return
		}

		// Try to parse the message
		var gameMsg GameMessage
		if err := json.Unmarshal(message, &gameMsg); err != nil {
			log.Println(fmt.Sprintf("[%s] Invalid message from %s", c.lobby.uuid[:10], c.user.Username))
			log.Println(fmt.Sprintf("[%s] %s", c.lobby.uuid[:10], err))
			continue
		}

		// Let the lobby handle the message
		c.lobby.message(c, gameMsg)
	}
}

// writeLoop pumps messages from the hub to the websocket connection.
func (c *Client) writeLoop() {
	log.Printf("[%s] Starting write loop for %s", c.lobby.uuid[:10], c.user.Username)
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
				log.Printf("[%s] Couldn't marshal the message for sending: %s", c.lobby.uuid[:10], err)
				return
			}

			if _, err = w.Write(messageBytes); err != nil {
				log.Printf("[%s] Couldn't write the message: %s", c.lobby.uuid[:10], err)
				return
			}

			// Add queued chat messages to the current websocket message.
			for i := 0; i < len(c.send); i++ {
				w.Write(newline)
				messageBytes, err := json.Marshal(<-c.send)
				if err != nil {
					log.Printf("[%s] Couldn't marshal the message for sending: %s", c.lobby.uuid[:10], err)
					return
				}

				if _, err = w.Write(messageBytes); err != nil {
					log.Printf("[%s] Couldn't write the message: %s", c.lobby.uuid[:10], err)
					return
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("[%s] Couldn't close the writer: %s", c.lobby.uuid[:10], err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[%s] Couldn't write the ping message: %s", c.lobby.uuid[:10], err)
				return
			}
		}
	}
}

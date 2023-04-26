package game

import (
	"encoding/json"
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
)

// msgType is the type of the game message.
type msgType string

const (
	MsgState  msgType = "state"
	MsgError  msgType = "error"
	MsgInput  msgType = "input"
	MsgAction msgType = "action"
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
		srv:   srv,
		conn:  conn,
		lobby: l,
		user:  user,
		send:  make(chan GameMessage, 256),
	}
}

// readLoop pumps messages from the websocket connection to the hub.
func (c *Client) readLoop() {
	defer func() {
		c.srv.unregisterQueue <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongWait)) })
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[%s] Connection aborted for %s: %v", c.lobby.uuid[:10], c.user.Username, err)
			}
			return
		}

		// Try to parse the message
		var gameMsg GameMessage
		if err := json.Unmarshal(message, &gameMsg); err != nil {
			log.Printf("[%s] Invalid message from %s", c.lobby.uuid[:10], c.user.Username)
			log.Printf("[%s] %s", c.lobby.uuid[:10], err)
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
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("[%s] Couldn't set the write deadline: %s", c.lobby.uuid[:10], err)
				return
			}

			if !ok {
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("[%s] Couldn't write the close message: %s", c.lobby.uuid[:10], err)
				}
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
				if _, err := w.Write(newline); err != nil {
					log.Printf("[%s] Couldn't write the newline: %s", c.lobby.uuid[:10], err)
					return
				}

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
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("[%s] Couldn't set the write deadline: %s", c.lobby.uuid[:10], err)
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[%s] Couldn't write the ping message: %s", c.lobby.uuid[:10], err)
				return
			}
		}
	}
}

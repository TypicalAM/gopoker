package websockets

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/chehsunliu/poker"
)

const gameUpdateInterval = 1 * time.Second

// Game is a game of poker played by multiple clients.
type Game struct {
	hub     *Hub
	UUID    string
	Players []*Client
	Deck    *poker.Deck
}

// newGame creates a new game with the given UUID and first player.
func newGame(hub *Hub, UUID string) *Game {
	return &Game{
		UUID: UUID,
		Deck: poker.NewDeck(),
	}
}

// run runs the game.
func (g *Game) run() {
	ticker := time.NewTicker(gameUpdateInterval)
	for {
		select {
		case <-ticker.C:
			//log.Printf("Game %s is ticking...", g.UUID)
		}
	}
}

// gameActionType is the type of game action.
type gameActionType string

const (
	actionDraw  gameActionType = "draw"
	actionFold                 = "fold"
	actionCall                 = "call"
	actionRaise                = "raise"
)

// gameAction is a game action sent by a client.
type gameAction struct {
	Type gameActionType
	Data string
}

// handleMessage handles a game message sent by a client. It returns a response and a boolean indicating
// if this response is meant for the sender only.
func (g *Game) handleMessage(client *Client, gameMsg *GameMessage) (response GameMessage, reply bool) {
	switch gameMsg.Type {
	case msgStart:
		return GameMessage{
			Type: msgStatus,
			Data: "Game already started, no need to start it yourself :)",
		}, true

	case msgStatus:
		return GameMessage{
			Type: msgStatus,
			Data: fmt.Sprintf("[%s] %s", client.player.Username, gameMsg.Data),
		}, false

	case msgAction:
		var action gameAction
		if err := json.Unmarshal([]byte(gameMsg.Data), &action); err != nil {
			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("Invalid action: %s", err),
			}, true
		}

		switch action.Type {
		case actionDraw:
			cardAmount, err := strconv.Atoi(action.Data)
			if err != nil {
				return GameMessage{
					Type: msgStatus,
					Data: fmt.Sprintf("Invalid card amount: %s", err),
				}, true
			}

			// We have to draw one card at a time because the guy who wrote the library
			// doesn't return the error from the draw method for some reason. The library was also updated
			// 5 years ago so I'm not sure if it's still maintained.
			var cards []poker.Card
			for i := 0; i < cardAmount; i++ {
				if g.Deck.Empty() {
					return GameMessage{
						Type: msgStatus,
						Data: fmt.Sprintf("[%s] tried to draw %d cards but the deck is empty", client.player.Username, cardAmount),
					}, false
				}
				cards = append(cards, g.Deck.Draw(1)...)
			}

			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("[%s] drew %s", client.player.Username, cards),
			}, false

		case actionFold:
			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("[%s] folded", client.player.Username),
			}, false

		case actionCall:
			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("[%s] called", client.player.Username),
			}, false

		case actionRaise:
			raiseAmount, err := strconv.Atoi(action.Data)
			if err != nil {
				return GameMessage{
					Type: msgStatus,
					Data: fmt.Sprintf("Invalid raise amount: %s", err),
				}, true
			}

			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("[%s] raised by %d", client.player.Username, raiseAmount),
			}, false

		default:
			return GameMessage{
				Type: msgStatus,
				Data: fmt.Sprintf("Invalid action type: %s", action.Type),
			}, true
		}

	case msgInput:
		return *gameMsg, false

	default:
		return GameMessage{
			Type: msgStatus,
			Data: fmt.Sprintf("Invalid message type: %s", gameMsg.Type),
		}, true
	}
}

// addPlayer adds a player to the game.
func (g *Game) addPlayer(player *Client) {
	g.Players = append(g.Players, player)
}

// removePlayer removes a player from the game.
func (g *Game) removePlayer(player *Client) {
	for i, p := range g.Players {
		if p == player {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			return
		}
	}
}

package websockets

import (
	"fmt"

	"github.com/chehsunliu/poker"
)

// Game is a game of poker played by multiple clients.
type Game struct {
	// ID is the unique identifier for the game
	UUID string

	// Players is a list of clients that are playing the game
	Players []*Client

	// Deck is the deck of cards that are used up in the game
	Deck *poker.Deck
}

// newGame creates a new game with the given UUID and first player.
func newGame(UUID string, firstPlayer *Client) *Game {
	return &Game{
		UUID:    UUID,
		Players: []*Client{firstPlayer},
		Deck:    poker.NewDeck(),
	}
}

// handleMessage handles a message from a client and returns a response.
func (g *Game) handleMessage(player *Client, message *GameMessage) *GameMessage {
	switch message.Type {
	case msgStart:
		return &GameMessage{
			Type: msgStatus,
			Data: "You aren't allowed to start the game!",
		}

	case msgAction:
		return &GameMessage{
			Type: msgStatus,
			Data: fmt.Sprintf("%s decided to draw 3 cards and drew %s", player.player.Username, g.Deck.Draw(3)),
		}

	case msgInput:
		// TODO: Handle input from the player
		return message

	default:
		return nil
	}
}

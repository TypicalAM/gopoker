package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
)

const gameUpdateInterval = 10 * time.Second

// Game is a game of poker played by multiple clients.
type Game struct {
	hub     *Hub
	UUID    string
	Players []*Client
	State   GameState
	StateMu sync.RWMutex
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
			g.checkWaiting()
		}
	}
}

// checkWaiting checks if the game is waiting for a player to make a move.
func (g *Game) checkWaiting() {
	g.StateMu.Lock()
	defer g.StateMu.Unlock()

	if !g.State.Started {
		return
	}

	if g.State.Waiting {
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("Waiting for %s to make their move", g.Players[g.State.Turn].player.Username))
	}
}

// startGame starts the game.
func (g *Game) startGame() {
	g.StateMu.Lock()
	defer g.StateMu.Unlock()

	// Create the game state.
	g.State = GameState{
		Started: true,
		Round:   preFlop,
		Waiting: true,
		Bets:    make([]int, len(g.Players)),
		Actions: make([]action, len(g.Players)),
		Assets:  make([]int, len(g.Players)),
		Hands:   make([][]poker.Card, len(g.Players)),
	}

	for i := range g.Players {
		g.State.Assets[i] = 100
	}

	g.sendToAllPlayers(msgStatus, "Game started")
	log.Println("Game started with", len(g.Players), "players")

	secondToLast := len(g.Players) - 2
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is the small blind", g.Players[secondToLast].player.Username))
	g.State.TotalBets = 1
	g.State.Bets[secondToLast] = 1
	g.State.Assets[secondToLast] -= 1

	log.Println("Players:", len(g.Players))
	last := len(g.Players) - 1
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is the big blind", g.Players[last].player.Username))
	g.State.TotalBets += 2
	g.State.Bets[last] = 2
	g.State.Assets[last] -= 2
	g.State.CurrentBet = 2

	for i := range g.Players {
		g.State.Hands[i] = g.Deck.Draw(2)
	}

	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is under the gun", g.Players[0].player.Username))
	g.sendToPlayer(0, msgStatus, "It's your turn")
	g.sendToPlayer(0, msgStatus, fmt.Sprintf("You have %d chips", g.State.Assets[0]))
	g.sendToPlayer(0, msgInput, "fold:call:raise")
	g.State.Waiting = true
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

// handleMessage handles a game message from a client and acts accordingly.
func (g *Game) handleMessage(client *Client, gameMsg *GameMessage) {
	index := -1
	for i, c := range g.Players {
		if c == client {
			index = i
			break
		}
	}

	if index == -1 {
		log.Printf("Client %s is not in game %s", client.player.Username, g.UUID)
		return
	}

	switch gameMsg.Type {
	case msgStart:
		g.sendToPlayer(index, msgStatus, "Game already started, no need to start it yourself :)")

	case msgStatus:
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] said %s", client.player.Username, gameMsg.Data))

	case msgAction:
		var action gameAction
		if err := json.Unmarshal([]byte(gameMsg.Data), &action); err != nil {
			g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid action: %s", err))
		}

		g.StateMu.Lock()
		defer g.StateMu.Unlock()

		switch action.Type {
		case actionDraw:
			cardAmount, err := strconv.Atoi(action.Data)
			if err != nil {
				g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid card amount: %s", err))
			}

			// We have to draw one card at a time because the guy who wrote the library
			// doesn't return the error from the draw method for some reason. The library was also updated
			// 5 years ago so I'm not sure if it's still maintained.
			var cards []poker.Card
			for i := 0; i < cardAmount; i++ {
				if g.Deck.Empty() {
					g.sendToPlayer(index, msgStatus, fmt.Sprintf("Tried to draw %d cards but the deck is empty", cardAmount))
				}

				cards = append(cards, g.Deck.Draw(1)...)
			}

			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] drew %s", client.player.Username, cards))

		case actionFold:
			if g.State.Waiting && g.State.Turn == index {
				g.State.Actions[index] = fold
				g.State.Turn++
				g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] folded", client.player.Username))
			} else {
				g.sendToPlayer(index, msgStatus, "It's not your turn")
			}

		case actionCall:
			if !g.State.Waiting || g.State.Turn != index {
				g.sendToPlayer(index, msgStatus, "It's not your turn")
				return
			}

			if g.State.Assets[index] < g.State.CurrentBet {
				g.sendToPlayer(index, msgStatus, fmt.Sprintf("You only have %d chips", g.State.Assets[index]))
				return
			}

			g.State.Actions[index] = call
			g.State.Assets[index] -= g.State.CurrentBet
			g.State.Bets[index] += g.State.CurrentBet
			g.State.TotalBets += g.State.CurrentBet
			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] called, they now have %d chips", client.player.Username, g.State.Assets[index]))
			g.State.Turn++

		case actionRaise:
			if !g.State.Waiting || g.State.Turn != index {
				g.sendToPlayer(index, msgStatus, "It's not your turn")
				return
			}

			raiseAmount, err := strconv.Atoi(action.Data)
			if err != nil {
				g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid raise amount: %s", err))
				return
			}

			if g.State.CurrentBet+raiseAmount > g.State.Assets[index] {
				g.sendToPlayer(index, msgStatus, fmt.Sprintf("You only have %d chips", g.State.Assets[index]))
				return
			}

			g.State.Actions[index] = raise
			g.State.Assets[index] -= g.State.CurrentBet + raiseAmount
			g.State.Bets[index] += g.State.CurrentBet + raiseAmount
			g.State.TotalBets += g.State.CurrentBet + raiseAmount
			g.State.CurrentBet += raiseAmount
			g.State.Turn++

			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] raised %d, they now have %d chips", client.player.Username, raiseAmount, g.State.Assets[index]))

		default:
			g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid action type: %s", action.Type))
		}

	case msgInput:
		// TODO: echoing is only a temporary solution, we need to implement a proper chat system. hehe
		g.sendToAllPlayers(msgInput, gameMsg.Data)

	default:
		g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid message type: %s", gameMsg.Type))
	}

	// Prompt the next player to take an action.
	if g.State.Turn < len(g.Players) {
		g.sendToPlayer(g.State.Turn, msgStatus, fmt.Sprintf("It's your turn, the current bet is %d", g.State.CurrentBet))

		// On the pre-flop we can't check
		if g.State.Round == preFlop {
			g.sendToPlayer(g.State.Turn, msgInput, "call:raise:fold")
		} else {
			g.sendToPlayer(g.State.Turn, msgInput, "check:call:raise:fold")
		}

		// Early return
		return
	}

	// If we reached this point, it means that all players have taken an action.
	if g.State.Round == river {
		// TODO: Implement the showdown.
		g.sendToAllPlayers(msgStatus, "The game is over!!!")
		g.sendToAllPlayers(msgStatus, "The game is over!!!")
		g.sendToAllPlayers(msgStatus, "The game is over!!!")
		g.State.Waiting = false
		return
	}

	// We need to move to the next round.
	g.State.Turn = 0
	g.State.CurrentBet = 2
	g.State.Round++

	g.sendToAllPlayers(msgStatus, "The round is over, moving to the next one...")

	// If someone folded, we need to remove them from the game.
	for i := 0; i < len(g.Players); i++ {
		if g.State.Actions[i] == fold {
			g.sendToPlayer(i, msgStatus, "You folded, you're out of the game")
			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] folded, they're out of the game", g.Players[i].player.Username))
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			g.State.Assets = append(g.State.Assets[:i], g.State.Assets[i+1:]...)
			g.State.Bets = append(g.State.Bets[:i], g.State.Bets[i+1:]...)
			g.State.Actions = append(g.State.Actions[:i], g.State.Actions[i+1:]...)
			i--
		}
	}

	// If there's only one player left, we need to end the game.
	if len(g.Players) == 1 {
		g.State.Assets[0] += g.State.TotalBets
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] won the game!!!", g.Players[0].player.Username))
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] has %d chips", g.Players[0].player.Username, g.State.Assets[0]))
		g.State.Waiting = false
		return
	}

	switch g.State.Round {
	case flop:
		g.sendToAllPlayers(msgStatus, "The flop is:")
		g.State.Cards = append(g.State.Cards, g.Deck.Draw(3)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.Cards))

	case turn:
		g.sendToAllPlayers(msgStatus, "The turn is:")
		g.State.Cards = append(g.State.Cards, g.Deck.Draw(1)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.Cards))

	case river:
		g.sendToAllPlayers(msgStatus, "The river is:")
		g.State.Cards = append(g.State.Cards, g.Deck.Draw(1)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.Cards))
	}

	// Make sure that the first player knows that it's their turn
	g.sendToPlayer(g.State.Turn, msgStatus, fmt.Sprintf("It's your turn, the current bet is %d", g.State.CurrentBet))
	g.sendToPlayer(g.State.Turn, msgInput, "call:raise:fold:check")
}

// addPlayer adds a player to the game.
func (g *Game) addPlayer(player *Client) {
	g.Players = append(g.Players, player)

	// TODO: Remove the hard-coded 3 players limit.
	if len(g.Players) == 3 {
		g.startGame()
	}
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

// sendToPlayer sends a message to a player.
func (g *Game) sendToPlayer(ind int, messageType msgType, msgData string) {
	log.Printf("[%s] game: %s", g.Players[ind].player.Username, msgData)
	msg := GameMessage{
		Type: messageType,
		Data: msgData,
	}

	select {
	case g.Players[ind].send <- msg:
	default:
		g.hub.unregister <- g.Players[ind]
	}
}

// sendToAllPlayers sends a message to all players.
func (g *Game) sendToAllPlayers(messageType msgType, msgData string) {
	log.Printf("[all] game: %s", msgData)
	msg := GameMessage{
		Type: messageType,
		Data: msgData,
	}

	for _, player := range g.Players {
		select {
		case player.send <- msg:
		default:
			g.hub.unregister <- player
		}
	}
}

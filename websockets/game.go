package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
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
	defer func() {
		g.broadcastState()
	}()

	if !g.State.Started {
		var flag bool
		if len(g.Players) == 3 {
			for i := 0; i < len(g.Players); i++ {
				log.Println(strings.TrimSpace(g.Players[i].userModel.UnsecuredCreditcard))
				log.Println(strings.TrimSpace(g.Players[i].userModel.UnsecuredCreditcard) == "")
			}

			for i := 0; i < len(g.Players); i++ {
				if strings.TrimSpace(g.Players[i].userModel.UnsecuredCreditcard) == "" {
					g.sendToPlayer(i, msgStatus, "You don't have chips to play, please input your credit card number")
					g.sendToPlayer(i, msgInput, "creditcard")
					flag = true
				}
			}

			log.Println("flag", flag)

			// If everyone has a credit card, we can start the game.
			if !flag {
				g.StateMu.Unlock()
				g.startGame()
				return
			}
		}
	}

	if g.State.Waiting {
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("Waiting for %s to make their move", g.Players[g.State.Turn].userModel.Username))
	}

	g.StateMu.Unlock()
}

// startGame starts the game.
func (g *Game) startGame() {
	g.StateMu.Lock()
	defer func() {
		g.broadcastState()
		g.StateMu.Unlock()
	}()

	// Create the game state.
	g.State = GameState{
		Started:        true,
		Round:          preFlop,
		Waiting:        true,
		Bets:           make([]int, len(g.Players)),
		Actions:        make([]action, len(g.Players)),
		Assets:         make([]int, len(g.Players)),
		Hands:          make([][]poker.Card, len(g.Players)),
		CommunityCards: []poker.Card{},
	}

	for i := range g.Players {
		g.State.Assets[i] = 100
	}

	g.sendToAllPlayers(msgStatus, "Game started")
	log.Println("Game started with", len(g.Players), "players")

	secondToLast := len(g.Players) - 2
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is the small blind", g.Players[secondToLast].userModel.Username))
	g.State.TotalBets = 1
	g.State.Bets[secondToLast] = 1
	g.State.Assets[secondToLast] -= 1

	log.Println("Players:", len(g.Players))
	last := len(g.Players) - 1
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is the big blind", g.Players[last].userModel.Username))
	g.State.TotalBets += 2
	g.State.Bets[last] = 2
	g.State.Assets[last] -= 2
	g.State.CurrentBet = 2

	for i := range g.Players {
		g.State.Hands[i] = g.Deck.Draw(2)
	}

	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] is under the gun", g.Players[0].userModel.Username))
	g.sendToPlayer(0, msgStatus, "It's your turn")
	g.sendToPlayer(0, msgStatus, fmt.Sprintf("You have %d chips", g.State.Assets[0]))
	g.sendToPlayer(0, msgInput, "fold:call:raise")
	g.State.Waiting = true
}

// gameActionType is the type of game action.
type gameActionType string

const (
	actionFold  gameActionType = "fold"
	actionCall                 = "call"
	actionRaise                = "raise"
	actionCheck                = "check"
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
		log.Printf("Client %s is not in game %s", client.userModel.Username, g.UUID)
		return
	}

	switch gameMsg.Type {
	case msgStart:
		g.sendToPlayer(index, msgStatus, "Game already started, no need to start it yourself :)")

	case msgStatus:
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] said %s", client.userModel.Username, gameMsg.Data))

	case msgAction:
		g.StateMu.Lock()
		defer func() {
			g.broadcastState()
			g.StateMu.Unlock()
		}()

		var action gameAction
		if err := json.Unmarshal([]byte(gameMsg.Data), &action); err != nil {
			g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid action: %s", err))
		}

		switch action.Type {
		case actionFold:
			if g.State.Waiting && g.State.Turn == index {
				g.State.Actions[index] = fold
				g.State.Turn++
				g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] folded", client.userModel.Username))
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
			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] called, they now have %d chips", client.userModel.Username, g.State.Assets[index]))
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

			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] raised %d, they now have %d chips", client.userModel.Username, raiseAmount, g.State.Assets[index]))

		case actionCheck:
			if !g.State.Waiting || g.State.Turn != index {
				g.sendToPlayer(index, msgStatus, "It's not your turn")
				return
			}

			for i := 0; i < index; i++ {
				if g.State.Actions[i] == call || g.State.Actions[i] == raise {
					g.sendToPlayer(index, msgStatus, "You can't check, someone before you has raised or called")
					return
				}
			}

			g.State.Actions[index] = check
			g.State.Turn++

		default:
			g.sendToPlayer(index, msgStatus, fmt.Sprintf("Invalid action type: %s", action.Type))
		}

	case msgInput:
		// If this is the credit input, we need to handle it differently.
		dataSplit := strings.Split(gameMsg.Data, ":")
		if len(dataSplit) != 2 || dataSplit[0] != "creditcard" {
			return
		}

		client.userModel.UnsecuredCreditcard = dataSplit[1]
		if resp := client.db.Save(client.userModel); resp.Error != nil {
			g.sendToPlayer(index, msgStatus, fmt.Sprintf("Error saving credit card: %s", resp.Error))
			return
		}

		log.Println("Credit card saved for user", client.userModel.Username)
		log.Println("Credit card:", client.userModel.UnsecuredCreditcard)
		return

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
		g.endGame()

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
			fmt.Println("Folded:", g.Players[i].userModel.Username)
			g.sendToPlayer(i, msgStatus, "You folded, you're out of the game")
			g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] folded, they're out of the game", g.Players[i].userModel.Username))
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			g.State.Assets = append(g.State.Assets[:i], g.State.Assets[i+1:]...)
			g.State.Bets = append(g.State.Bets[:i], g.State.Bets[i+1:]...)
			g.State.Actions = append(g.State.Actions[:i], g.State.Actions[i+1:]...)
			i--
		}
	}

	g.State.Actions = make([]action, len(g.Players))
	g.State.Bets = make([]int, len(g.Players))

	// If there's only one player left, we need to end the game.
	if len(g.Players) == 1 {
		g.State.Assets[0] += g.State.TotalBets
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] won the game!!!", g.Players[0].userModel.Username))
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] has %d chips", g.Players[0].userModel.Username, g.State.Assets[0]+g.State.TotalBets))
		g.State.Waiting = false

		for i := 0; i < len(g.Players); i++ {
			g.State.Usernames = append(g.State.Usernames, g.Players[i].userModel.Username)
		}

		stateBytes, _ := json.Marshal(g.State)
		g.sendToAllPlayers(msgState, string(stateBytes))
		g.sendToAllPlayers(msgGameEnd, fmt.Sprintf("%d:%s", 0, ""))
		return
	}

	switch g.State.Round {
	case flop:
		g.sendToAllPlayers(msgStatus, "The flop is:")
		g.State.CommunityCards = append(g.State.CommunityCards, g.Deck.Draw(3)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.CommunityCards))

	case turn:
		g.sendToAllPlayers(msgStatus, "The turn is:")
		g.State.CommunityCards = append(g.State.CommunityCards, g.Deck.Draw(1)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.CommunityCards))

	case river:
		g.sendToAllPlayers(msgStatus, "The river is:")
		g.State.CommunityCards = append(g.State.CommunityCards, g.Deck.Draw(1)...)
		g.sendToAllPlayers(msgStatus, fmt.Sprintf("%s", g.State.CommunityCards))
	}

	// Make sure that the first player knows that it's their turn
	g.sendToPlayer(g.State.Turn, msgStatus, fmt.Sprintf("It's your turn, the current bet is %d", g.State.CurrentBet))
	g.sendToPlayer(g.State.Turn, msgInput, "call:raise:fold:check")
}

// addPlayer adds a player to the game.
func (g *Game) addPlayer(player *Client) {
	g.Players = append(g.Players, player)

	// TODO: Remove the hard-coded 3 players limit.
	var flag bool
	if len(g.Players) == 3 {
		for i := 0; i < len(g.Players); i++ {
			if strings.TrimSpace(g.Players[i].userModel.UnsecuredCreditcard) == "" {
				g.sendToPlayer(i, msgStatus, "You don't have chips to play, please input your credit card number")
				g.sendToPlayer(i, msgInput, "creditcard")
				flag = true
			}
		}
		// If everyone has a credit card, we can start the game.
		if !flag {
			g.startGame()
		}
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
	if messageType != msgState {
		log.Printf("[%s] game: %s", g.Players[ind].userModel.Username, msgData)
	}

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

// broadcastState sends a periodic "state" message to all players in the game. This is used to keep the players
// in sync with the game state.
func (g *Game) broadcastState() {
	if !g.State.Started {
		return
	}

	safeState := g.State
	safeHands := make([][]poker.Card, len(g.State.Hands))
	copy(safeHands, g.State.Hands)

	// Make sure that this field is not copied
	safeState.Hands = make([][]poker.Card, len(g.State.Hands))
	for i := range g.State.Hands {
		safeState.Hands[i] = []poker.Card{}
	}

	// Show the usernames of the players.
	safeState.Usernames = []string{}
	for _, player := range g.Players {
		safeState.Usernames = append(safeState.Usernames, player.userModel.Username)
	}

	// Show the cards of the players that are still in the game.
	for i := range g.Players {
		safeState.Hands[i] = safeHands[i]

		stateBytes, _ := json.Marshal(safeState)
		g.sendToPlayer(i, msgState, string(stateBytes))

		safeState.Hands[i] = []poker.Card{}
	}
}

// determineWinner determines the winner of the game.
func (g *Game) endGame() {
	bestHand := make([]poker.Card, 5)
	bestRank := ""
	bestScore := math.MaxInt32
	bestPlayer := -1
	for i := range g.Players {
		hand, score, rank := getBestHand(g.State.Hands[i], g.State.CommunityCards)
		if score < bestScore {
			bestScore = score
			bestRank = rank
			bestHand = hand
			bestPlayer = i
		}
	}

	g.State.Waiting = false
	g.State.Usernames = []string{}
	for _, player := range g.Players {
		g.State.Usernames = append(g.State.Usernames, player.userModel.Username)
	}

	stateBytes, _ := json.Marshal(g.State)
	g.sendToAllPlayers(msgState, string(stateBytes))

	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] won the game!!!", g.Players[bestPlayer].userModel.Username))
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("The best rank is %s", bestRank))
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("The best hand is %s", bestHand))
	g.sendToAllPlayers(msgStatus, fmt.Sprintf("[%s] has %d chips", g.Players[bestPlayer].userModel.Username, g.State.Assets[bestPlayer]+g.State.TotalBets))
	g.sendToAllPlayers(msgGameEnd, fmt.Sprintf("%d:%s", bestPlayer, bestRank))
}

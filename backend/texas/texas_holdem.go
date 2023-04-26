package texas

import (
	"errors"
	"math"

	"github.com/chehsunliu/poker"
	"gonum.org/v1/gonum/stat/combin"
)

type pokerRound string

const (
	PreFlop pokerRound = "preflop"
	Flop    pokerRound = "flop"
	Turn    pokerRound = "turn"
	River   pokerRound = "river"
)

type PokerAction string

const (
	None  PokerAction = "none"
	Call  PokerAction = "call"
	Raise PokerAction = "raise"
	Check PokerAction = "check"
	Fold  PokerAction = "fold"
)

var actionMap = map[string]PokerAction{
	"none":  None,
	"call":  Call,
	"raise": Raise,
	"check": Check,
	"fold":  Fold,
}

var NotEnoughMoneyErr = errors.New("Not enough money")
var WrongTurnErr = errors.New("Wrong turn")
var InvalidActionErr = errors.New("Wrong action")
var NotEnoughPlayersErr = errors.New("Not enough players")
var GameStillInProgressErr = errors.New("Game still in progress")
var PlayerNotInGameErr = errors.New("Player not in game")
var OwnTurnDisconnectErr = errors.New("Own turn disconnection")
var DrawErr = errors.New("Internal error")
var InvalidAssetErr = errors.New("Invalid asset number")
var InternalErr = errors.New("Internal error")

const RequiredPlayers = 3

type TexasHoldEm struct {
	deck           *poker.Deck
	CommunityCards []poker.Card
	Players        []Player
	Round          pokerRound
	CurrentPlayer  int
	ActiveBet      int
	Pot            int

	gameStarted bool
	GameOver    bool
	GameWinner  string
	BestRank    string
	BestHand    []poker.Card
}

type Player struct {
	Name      string
	HoleCards []poker.Card
	Assets    int
	Bet       int
	Action    PokerAction
	Active    bool
}

func NewTexasHoldEm() *TexasHoldEm {
	return &TexasHoldEm{
		deck: poker.NewDeck(),
	}
}

func (t *TexasHoldEm) AddPlayer(username string, assets int) error {
	if assets < 0 {
		return InvalidAssetErr
	}

	for _, player := range t.Players {
		if player.Name == username {
			// This means the player is reconnecting, so we don't want to add them again
			// but they should be folding for the rest of the game
			return nil
		}
	}

	t.Players = append(t.Players, Player{
		Name:   username,
		Assets: assets,
	})

	return nil
}

func (t *TexasHoldEm) StartGame() error {
	if t.gameStarted {
		return GameStillInProgressErr
	}

	if len(t.Players) < RequiredPlayers {
		return NotEnoughPlayersErr
	}

	t.gameStarted = true
	t.Round = PreFlop
	for i := range t.Players {
		cards, err := safeDraw(t.deck, 2)
		if err != nil {
			return err
		}

		t.Players[i].HoleCards = make([]poker.Card, 2)
		copy(t.Players[i].HoleCards, cards)
		t.Players[i].Active = true
		t.Players[i].Bet = 0
	}

	bigBlind := len(t.Players) - 1
	if t.Players[bigBlind].Assets < 2 {
		return NotEnoughMoneyErr
	}

	t.Players[bigBlind].Bet = 2
	t.Players[bigBlind].Assets -= 2

	smallBlind := bigBlind - 1
	if t.Players[smallBlind].Assets < 1 {
		return NotEnoughMoneyErr
	}

	t.Players[smallBlind].Bet = 1
	t.Players[smallBlind].Assets -= 1

	t.Pot = 3
	t.ActiveBet = 2

	return nil
}

func (t *TexasHoldEm) AdvanceState(username string, action PokerAction) error {
	if len(t.Players) < RequiredPlayers {
		return NotEnoughPlayersErr
	}

	playerIndex := -1
	for i, player := range t.Players {
		if player.Name == username {
			playerIndex = i
		}
	}

	if playerIndex == -1 {
		return PlayerNotInGameErr
	}

	if playerIndex != t.CurrentPlayer {
		return WrongTurnErr
	}

	if action == None || !t.Players[playerIndex].Active {
		return InvalidActionErr
	}

	switch action {
	case Call:
		if t.Players[playerIndex].Assets < t.ActiveBet {
			return NotEnoughMoneyErr
		}

		t.Pot += t.ActiveBet - t.Players[playerIndex].Bet
		t.Players[playerIndex].Assets -= t.ActiveBet - t.Players[playerIndex].Bet
		t.Players[playerIndex].Bet = t.ActiveBet
		t.Players[playerIndex].Action = Call

	case Raise:
		if t.Players[playerIndex].Assets < t.ActiveBet+2 {
			return NotEnoughMoneyErr
		}

		t.Pot += t.ActiveBet - t.Players[playerIndex].Bet + 2
		t.Players[playerIndex].Assets -= t.ActiveBet + 2 - t.Players[playerIndex].Bet
		t.Players[playerIndex].Bet = t.ActiveBet + 2
		t.Players[playerIndex].Action = Raise

	case Check:
		if t.Round == PreFlop {
			return InvalidActionErr
		}

		if playerIndex != 0 {
			for i := playerIndex - 1; i >= 0; i-- {
				if t.Players[i].Action != Check {
					return InvalidActionErr
				}
			}
		}

		t.Players[playerIndex].Action = Check

	case Fold:
		t.Players[playerIndex].Action = Fold
		t.Players[playerIndex].Active = false
	}

	var looped bool
	t.CurrentPlayer, looped = t.getNextPlayer(t.CurrentPlayer)
	if !looped {
		return nil
	}

	switch t.Round {
	case PreFlop:
		cards, err := safeDraw(t.deck, 3)
		if err != nil {
			return err
		}

		t.CommunityCards = append(t.CommunityCards, cards...)
		t.Round = Flop

	case Flop:
		card, err := safeDraw(t.deck, 1)
		if err != nil {
			return err
		}

		t.CommunityCards = append(t.CommunityCards, card[0])
		t.Round = Turn

	case Turn:
		card, err := safeDraw(t.deck, 1)
		if err != nil {
			return err
		}

		t.CommunityCards = append(t.CommunityCards, card[0])
		t.Round = River

	case River:
		t.GameOver = true
		winner, rank, hand, err := t.getWinner()
		t.GameWinner = winner
		t.BestRank = rank
		t.BestHand = hand
		return err
	}

	playersActive := 0
	lastActivePlayer := -1
	for i, player := range t.Players {
		if player.Active {
			playersActive++
			lastActivePlayer = i
		}
	}

	if playersActive == 1 {
		t.GameOver = true
		t.GameWinner = t.Players[lastActivePlayer].Name
		t.BestRank = "Last man standing"
		return nil
	}

	for i := range t.Players {
		if t.Players[i].Active {
			t.Players[i].Action = None
		}
	}

	smallBlind, _ := t.getNextPlayer(0)
	if t.Players[smallBlind].Assets < 1 {
		return NotEnoughMoneyErr
	}

	t.Players[smallBlind].Bet = 1
	t.Players[smallBlind].Assets -= 1

	bigBlind, _ := t.getNextPlayer(smallBlind)
	if t.Players[bigBlind].Assets < 2 {
		return NotEnoughMoneyErr
	}

	t.Players[bigBlind].Bet = 2
	t.Players[bigBlind].Assets -= 2

	t.ActiveBet = 2
	return nil
}

func (t *TexasHoldEm) getNextPlayer(current int) (int, bool) {
	numPlayers := len(t.Players)
	looped := false
	for i := 1; i < numPlayers; i++ {
		index := (current + i) % numPlayers
		if index == 0 {
			looped = true
		}

		if t.Players[index].Active {
			return index, looped
		}

		if index == numPlayers-1 {
			looped = true
		}
	}

	return -1, looped
}

func (t TexasHoldEm) SanitizeState(username string) *TexasHoldEm {
	if t.GameOver {
		return &t
	}

	sanitized := t
	sanitized.Players = make([]Player, len(t.Players))
	for i, player := range t.Players {
		sanitized.Players[i] = player
		if player.Name != username {
			sanitized.Players[i].HoleCards = []poker.Card{}
		}
	}

	return &sanitized
}

func (t *TexasHoldEm) Disconnect(username string) error {
	var found bool
	var index int
	for i, player := range t.Players {
		if player.Name == username {
			found = true
			index = i
		}
	}

	if !found {
		return PlayerNotInGameErr
	}

	if index == t.CurrentPlayer {
		if err := t.AdvanceState(username, Fold); err != nil {
			return err
		}

		return OwnTurnDisconnectErr
	}

	if len(t.Players) == 1 {
		t.GameOver = true
		t.GameWinner = t.Players[0].Name
		t.BestRank = "Last man standing..."
		return nil
	}

	if len(t.Players) == 0 {
		t.GameOver = true
		t.GameWinner = "No one"
		t.BestRank = "No one won..."
		return nil
	}

	return nil
}

func (t *TexasHoldEm) IsGameOver() bool {
	return t.GameOver
}

func (t *TexasHoldEm) getWinner() (string, string, []poker.Card, error) {
	if !t.GameOver {
		return "", "", []poker.Card{}, GameStillInProgressErr
	}

	var bestRank string
	var bestHand []poker.Card
	var bestPlayer string
	bestScore := math.MaxInt32

	for _, player := range t.Players {
		if !player.Active {
			continue
		}

		hand, score, rank := getBestHand(player.HoleCards, t.CommunityCards)
		if score < bestScore {
			bestScore = score
			bestRank = rank
			bestHand = hand
			bestPlayer = player.Name
		}
	}

	return bestPlayer, bestRank, bestHand, nil
}

func (t *TexasHoldEm) ShouldBeDisbanded() bool {
	return t.GameOver || (t.gameStarted && len(t.Players) == 0)
}

func DecodeAction(text string) (PokerAction, bool) {
	val, ok := actionMap[text]
	return val, ok
}

func safeDraw(deck *poker.Deck, n int) ([]poker.Card, error) {
	if deck == nil {
		return []poker.Card{}, DrawErr
	}

	cards := make([]poker.Card, n)
	for i := 0; i < n; i++ {
		if deck.Empty() {
			return []poker.Card{}, DrawErr
		}

		cards[i] = deck.Draw(1)[0]
	}

	return cards, nil
}

func getBestHand(holeCards []poker.Card, communityCards []poker.Card) ([]poker.Card, int, string) {
	combinedCards := append(holeCards, communityCards...)
	bestHand := make([]poker.Card, 5)
	bestScore := int32(math.MaxInt32)
	var bestRank string
	permgen := combin.NewPermutationGenerator(7, 5)
	currentHand := make([]poker.Card, 5)

	for permgen.Next() {
		hand := permgen.Permutation(nil)
		for i := 0; i < 5; i++ {
			currentHand[i] = combinedCards[hand[i]]
		}

		score := poker.Evaluate(currentHand)
		if score < bestScore {
			bestScore = score
			copy(bestHand, currentHand)
			bestRank = poker.RankString(score)
		}
	}

	return bestHand, int(bestScore), bestRank
}

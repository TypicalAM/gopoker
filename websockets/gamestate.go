package websockets

import "github.com/chehsunliu/poker"

// round is the round of the game.
type round int

const (
	preFlop round = iota
	flop
	turn
	river
)

// action is the action of a player.
type action int

const (
	none action = iota
	fold
	check
	call
	raise
)

// GameState is the state of the game.
type GameState struct {
	Started bool
	Round   round
	Waiting bool

	Cards      []poker.Card
	CurrentBet int
	Turn       int
	TotalBets  int
	Bets       []int
	Actions    []action
	Assets     []int
	Hands      [][]poker.Card
}

package game

import (
	"errors"
	"fmt"
	"testing"

	"github.com/chehsunliu/poker"
)

// testGame creates a test game.
func testGame() *TexasHoldEm {
	texas := NewTexasHoldEm()
	texas.AddPlayer("Player 0", 100)
	texas.AddPlayer("Player 1", 100)
	texas.AddPlayer("Player 2", 100)
	texas.StartGame()
	return texas
}

// TestDetermineWinner tests the getWinner function.
func TestDetermineWinner(t *testing.T) {
	tt := []struct {
		name     string
		comunity []poker.Card
		hole0    []poker.Card
		hole1    []poker.Card
		hole2    []poker.Card
		bestRank string
		winner   string
	}{
		{
			name: "p0: four of a kind",
			comunity: []poker.Card{
				poker.NewCard("Qs"),
				poker.NewCard("Qh"),
				poker.NewCard("Ts"),
				poker.NewCard("9s"),
				poker.NewCard("8s"),
			},
			hole0:    []poker.Card{poker.NewCard("Qd"), poker.NewCard("Qc")},
			hole1:    []poker.Card{poker.NewCard("2h"), poker.NewCard("2s")},
			hole2:    []poker.Card{poker.NewCard("3h"), poker.NewCard("3s")},
			bestRank: "Four of a Kind",
			winner:   "Player 0",
		},
		{
			name: "p1: full house",
			comunity: []poker.Card{
				poker.NewCard("Qs"),
				poker.NewCard("Qh"),
				poker.NewCard("Qc"),
				poker.NewCard("9s"),
				poker.NewCard("8s"),
			},
			hole0:    []poker.Card{poker.NewCard("2h"), poker.NewCard("3s")},
			hole1:    []poker.Card{poker.NewCard("2d"), poker.NewCard("2c")},
			hole2:    []poker.Card{poker.NewCard("3h"), poker.NewCard("4s")},
			bestRank: "Full House",
			winner:   "Player 1",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			texas := testGame()
			texas.CommunityCards = tc.comunity
			texas.Players[0].HoleCards = tc.hole0
			texas.Players[1].HoleCards = tc.hole1
			texas.Players[2].HoleCards = tc.hole2
			texas.gameOver = true

			player, rank, hand, err := texas.getWinner()
			if err != nil {
				t.Fatal(err)
			}

			if player != tc.winner {
				t.Errorf("%v", hand)
				t.Errorf("expected winner to be %s, got %s", tc.winner, player)
			}

			if rank != tc.bestRank {
				t.Errorf("expected best rank to be %s, got %s", tc.bestRank, rank)
			}
		})
	}
}

// TestSanitizeState tests the state sanitization function.
func TestSanitizeState(t *testing.T) {
	tt := []struct {
		name    string
		current int
	}{
		{
			name:    "player 0",
			current: 0,
		},
		{
			name:    "player 1",
			current: 1,
		},
		{
			name:    "player 2",
			current: 2,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			texas := testGame()
			sanitized := texas.SanitizeState(texas.Players[tc.current].Name)
			for i, player := range sanitized.Players {
				if i == tc.current {
					if len(player.HoleCards) == 0 {
						t.Errorf("expected hole cards to have a non-zero length, got zero length")
					}
				} else {
					if len(player.HoleCards) != 0 {
						t.Errorf("expected hole cards to have a zero length, got non-zero length")
					}
				}
			}
		})
	}
}

// testPlayerAction is a struct used to automate player actions.
type testPlayerAction struct {
	player string
	action pokerAction
}

// handlePlayerAction handles a player action.
func handlePlayerActions(texas *TexasHoldEm, moves []testPlayerAction) error {
	for _, move := range moves {
		err := texas.AdvanceState(move.player, move.action)
		if err != nil {
			return err
		}
	}

	return nil
}

// TestNormalGame tests a normal game of Texas Hold'em with no folds.
func TestNormalGame(t *testing.T) {
	texas := testGame()
	threeCalls := make([]testPlayerAction, 3)
	for i := 0; i < 3; i++ {
		threeCalls[i] = testPlayerAction{
			player: fmt.Sprintf("Player %d", i),
			action: Call,
		}
	}

	if err := handlePlayerActions(texas, threeCalls); err != nil {
		t.Fatal(err)
	}

	if texas.Round != Flop {
		t.Errorf("expected current state to be flop, got %s", texas.Round)
	}

	if len(texas.CommunityCards) != 3 {
		t.Errorf("expected community cards to have a length of 3, got %d", len(texas.CommunityCards))
	}

	if err := handlePlayerActions(texas, threeCalls); err != nil {
		t.Fatal(err)
	}

	if texas.Round != Turn {
		t.Errorf("expected current state to be turn, got %s", texas.Round)
	}

	if len(texas.CommunityCards) != 4 {
		t.Errorf("expected community cards to have a length of 4, got %d", len(texas.CommunityCards))
	}

	if err := handlePlayerActions(texas, threeCalls); err != nil {
		t.Fatal(err)
	}

	if texas.Round != River {
		t.Errorf("expected current state to be river, got %s", texas.Round)
	}

	if len(texas.CommunityCards) != 5 {
		t.Errorf("expected community cards to have a length of 5, got %d", len(texas.CommunityCards))
	}

	if err := handlePlayerActions(texas, threeCalls); err != nil {
		t.Fatal(err)
	}

	if !texas.IsGameOver() {
		t.Errorf("expected game to be over, got not over")
	}
}

// TestFolds tests a game of Texas Hold'em with folds.
func TestFolds(t *testing.T) {
	moves := []testPlayerAction{
		{"Player 0", Fold},
		{"Player 1", Call},
		{"Player 2", Call},
		{"Player 0", Call},
	}

	if err := handlePlayerActions(testGame(), moves); err == nil || !errors.Is(err, WrongTurnErr) {
		t.Errorf("expected wrong turn error, got %v", err)
	}

	moves = []testPlayerAction{
		{"Player 0", Call},
		{"Player 1", Fold},
		{"Player 2", Fold},
	}

	texas := testGame()
	if err := handlePlayerActions(texas, moves); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !texas.IsGameOver() {
		t.Errorf("expected game to be over, got not over")
	}

	if texas.GameWinner != "Player 0" {
		t.Errorf("expected player 0 to win, got %s", texas.GameWinner)
	}

	moves = []testPlayerAction{
		{"Player 0", Fold},
		{"Player 1", Fold},
		{"Player 2", Call},
	}

	texas = testGame()
	if err := handlePlayerActions(texas, moves); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !texas.IsGameOver() {
		t.Errorf("expected game to be over, got not over")
	}

	if texas.GameWinner != "Player 2" {
		t.Errorf("expected player 2 to win, got %s", texas.GameWinner)
	}
}

// TestIllegalActions tests a game of Texas Hold'em with illegal actions.
func TestIllegalActions(t *testing.T) {
	moves := []testPlayerAction{
		{"Player 0", Call},
		{"Player 1", Check},
	}

	if err := handlePlayerActions(testGame(), moves); err == nil || !errors.Is(err, InvalidActionErr) {
		t.Errorf("expected illegal action error, got %v", err)
	}

	moves = []testPlayerAction{
		{"Player 0", Call},
		{"Player 0", Check},
	}

	if err := handlePlayerActions(testGame(), moves); err == nil || !errors.Is(err, WrongTurnErr) {
		t.Errorf("expected illegal action error, got %v", err)
	}

	moves = []testPlayerAction{
		{"Player 0", Call},
		{"Player 1", Call},
		{"Player 2", Call},
		{"Player 0", Call},
		{"Player 1", Check},
	}

	if err := handlePlayerActions(testGame(), moves); err == nil || !errors.Is(err, InvalidActionErr) {
		t.Errorf("expected illegal action error, got %v", err)
	}
}

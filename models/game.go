package models

import "gorm.io/gorm"

// GameIDKey is the key for the game ID in the session
var GameIDKey = "gameID"

// Game represents a game of poker
type Game struct {
	gorm.Model
	Playing bool
	UUID    string
	Players []User
}

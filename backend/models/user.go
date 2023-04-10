package models

import "gorm.io/gorm"

// User holds information about a user.
type User struct {
	gorm.Model
	Username            string
	Password            string
	GameID              uint
	// TODO: This is purposefully not secure, it's just to show that we can
	// sniff it out of the websocket connection.
	UnsecuredCreditcard string
	Sessions            []Session
}

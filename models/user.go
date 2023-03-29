package models

import "gorm.io/gorm"

// User holds information about a user.
type User struct {
	gorm.Model
	Username string
	Password string
	Sessions []Session
}

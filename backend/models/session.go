package models

import (
	"time"

	"gorm.io/gorm"
)

// Session holds information about user sesstions and when they expire
type Session struct {
	gorm.Model
	Identifier string
	UserID     uint
	ExpiresAt  time.Time
}

// HasExpired returns true if the session has expired
func (s *Session) HasExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

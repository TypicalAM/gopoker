package models

import (
	"log"

	"gorm.io/gorm"
)

// User holds information about a user.
type User struct {
	gorm.Model
	Username string
	Password string
	GameID   *uint
	Profile  Profile
	Sessions []Session
}

// Profile holds information about a user's profile.
type Profile struct {
	UserID      uint
	DisplayName string
	ImageURL    string
}

// AfterCreate is a hook that is called to make sure that a profile is created for the user.
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	profile := Profile{
		UserID:      u.ID,
		DisplayName: u.Username,
		// TODO: Get default image URL from config
		ImageURL: "https://www.stockvault.net/data/2009/07/28/109653/preview16.jpg",
	}

	if res := tx.Create(&profile); res.Error != nil {
		log.Println("Error creating profile", res.Error)
		return res.Error
	}

	u.Profile = profile
	return nil
}

// SafeUser is a safe user representation.
type SafeUser struct {
	Username string      `json:"username"`
	Profile  SafeProfile `json:"profile"`
}

// SafeProfile is a safe profile representation.
type SafeProfile struct {
	DisplayName string `json:"display_name"`
	ImageURL    string `json:"image_url"`
}

// Sanitize returns a safe user representation.
func (u *User) Sanitize() SafeUser {
	return SafeUser{
		Username: u.Username,
		Profile: SafeProfile{
			DisplayName: u.Profile.DisplayName,
			ImageURL:    u.Profile.ImageURL,
		},
	}
}

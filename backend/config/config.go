package config

import (
	"os"
	"strconv"
	"strings"
)

// Config is a struct that holds the configuration for the application.
type Config struct {
	DatabaseUser       string
	DatabasePassword   string
	DatabaseHost       string
	DatabasePort       string
	DatabaseName       string
	DatabaseTestName   string
	CookieSecret       string
	RequestsPerMin     int
	ListenPort         string
	GamePlayerCap      int
	CorsTrustedOrigins []string

	// TODO: Make this optional
	CloudinaryURL string
}

// ReadConfig reads the config from the .env file and populates the Config struct.
func ReadConfig() (*Config, error) {
	requestsPerMinRaw := os.Getenv("REQUESTS_PER_MIN")
	requestsPerMin, err := strconv.Atoi(requestsPerMinRaw)
	if err != nil {
		return nil, err
	}

	gameplayercapRaw := os.Getenv("GAME_PLAYER_CAP")
	gameplayercap, err := strconv.Atoi(gameplayercapRaw)
	if err != nil {
		return nil, err
	}

	corsTrustedOriginsRaw := os.Getenv("CORS_TRUSTED_ORIGINS")
	corsTrustedOrigins := strings.Split(corsTrustedOriginsRaw, ",")

	cloudinaryURL := os.Getenv("CLOUDINARY_URL")

	cfg := &Config{
		DatabaseUser:       os.Getenv("DB_USER"),
		DatabasePassword:   os.Getenv("DB_PASSWORD"),
		DatabaseHost:       os.Getenv("DB_HOST"),
		DatabasePort:       os.Getenv("DB_PORT"),
		DatabaseName:       os.Getenv("DB_DATABASE"),
		DatabaseTestName:   os.Getenv("DB_TEST_DATABASE"),
		CookieSecret:       os.Getenv("COOKIE_SECRET"),
		RequestsPerMin:     requestsPerMin,
		ListenPort:         os.Getenv("LISTEN_PORT"),
		GamePlayerCap:      gameplayercap,
		CorsTrustedOrigins: corsTrustedOrigins,
		CloudinaryURL:      cloudinaryURL,
	}

	return cfg, nil
}

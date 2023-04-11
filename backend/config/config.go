package config

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config is a struct that holds the configuration for the application.
type Config struct {
	MySQLUser         string
	MySQLPassword     string
	MySQLHost         string
	MySQLPort         string
	MySQLDatabase     string
	MySQLTestDatabase string
	CookieSecret      string
	CacheLifetime     int
	CacheParameter    string
	RequestsPerMin    int
	ListenPort        string
	GamePlayerCap     int
}

// ReadConfig reads the config from the .env file and populates the Config struct.
func ReadConfig(dir string) (*Config, error) {
	err := godotenv.Load(filepath.Join(dir, ".env"))
	if err != nil {
		return nil, err
	}

	cacheLifetimeRaw := os.Getenv("CACHE_LIFETIME")
	cacheLifetime, err := strconv.Atoi(cacheLifetimeRaw)
	if err != nil {
		return nil, err
	}

	cacheParameter := os.Getenv("CACHE_PARAMETER")
	if cacheParameter == "" {
		var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		cacheParameter = ""
		for i := 0; i < 10; i++ {
			cacheParameter += string(letters[rand.Intn(len(letters))])
		}
	}

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

	cfg := &Config{
		MySQLUser:         os.Getenv("MYSQL_USER"),
		MySQLPassword:     os.Getenv("MYSQL_PASSWORD"),
		MySQLHost:         os.Getenv("MYSQL_HOST"),
		MySQLPort:         os.Getenv("MYSQL_PORT"),
		MySQLDatabase:     os.Getenv("MYSQL_DATABASE"),
		MySQLTestDatabase: os.Getenv("MYSQL_TEST_DATABASE"),
		CookieSecret:      os.Getenv("COOKIE_SECRET"),
		CacheLifetime:     cacheLifetime,
		CacheParameter:    cacheParameter,
		RequestsPerMin:    requestsPerMin,
		ListenPort:        os.Getenv("LISTEN_PORT"),
		GamePlayerCap:     gameplayercap,
	}

	return cfg, nil
}

package config

import (
	"os"
	"strconv"
	"strings"
)

// FileUploadService is an enum for the file upload service.
type FileUploadService int

const (
	Cloudinary FileUploadService = iota
	Local
)

var fileUploadMap = map[string]FileUploadService{
	"cloudinary": Cloudinary,
	"local":      Local,
}

// Config is a struct that holds the configuration for the application.
type Config struct {
	// Database related
	DatabaseUser     string
	DatabasePassword string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string

	// Game related
	GamePlayerCap int

	// Server related
	ListenPort         string
	CookieSecret       string
	RequestsPerMin     int
	CorsTrustedOrigins []string

	// Upload related
	FileUploadType FileUploadService
	CloudinaryURL  string
	FileUploadPath string
}

// New returns a new Config struct.
func New() *Config {
	return &Config{
		DatabaseUser:       getEnvString("DB_USER", "myuser"),
		DatabasePassword:   getEnvString("DB_PASSWORD", "mypassword"),
		DatabaseHost:       getEnvString("DB_HOST", "localhost"),
		DatabasePort:       getEnvString("DB_PORT", "5432"),
		DatabaseName:       getEnvString("DB_DATABASE", "mydatabase"),
		CookieSecret:       getEnvString("COOKIE_SECRET", "mysecret"),
		RequestsPerMin:     getEnvInt("REQUESTS_PER_MIN", 30),
		ListenPort:         getEnvString("LISTEN_PORT", "8080"),
		GamePlayerCap:      getEnvInt("GAME_PLAYER_CAP", 3),
		CorsTrustedOrigins: strings.Split(getEnvString("CORS_TRUSTED_ORIGINS", "http://localhost:3000"), ","),
		FileUploadType:     getEnvFileUpload("FILE_UPLOAD_TYPE", Local),
		CloudinaryURL:      getEnvString("CLOUDINARY_URL", ""),
		FileUploadPath:     getEnvString("FILE_UPLOAD_PATH", "uploads"),
	}
}

// NewTest returns a new Config struct for testing.
func NewTest() *Config {
	return &Config{
		DatabaseUser:       getEnvString("DB_USER", "myuser"),
		DatabasePassword:   getEnvString("DB_PASSWORD", "mypassword"),
		DatabaseHost:       getEnvString("DB_TEST_HOST", "localhost"),
		DatabasePort:       getEnvString("DB_PORT", "5432"),
		DatabaseName:       getEnvString("DB_TEST_DATABASE", "mytestdatabase"),
		CookieSecret:       "cokkie",
		RequestsPerMin:     1000,
		ListenPort:         "8080",
		GamePlayerCap:      3,
		CorsTrustedOrigins: strings.Split(getEnvString("CORS_TRUSTED_ORIGINS", "http://localhost:3000"), ","),
		CloudinaryURL:      getEnvString("CLOUDINARY_URL", ""),
	}
}

// getEnvString gets the environment variable or returns the default value.
func getEnvString(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

// getEnvInt gets the environment variable or returns the default value.
func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	num, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return num
}

// getEnvFileUpload returns the file upload service.
func getEnvFileUpload(key string, fallback FileUploadService) FileUploadService {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	service, ok := fileUploadMap[val]
	if !ok {
		return fallback
	}

	return service
}

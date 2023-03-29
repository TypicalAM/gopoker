package gopoker

import (
	"log"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ConnectToDatabase connects to the database using the config.
func ConnectToDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.MySQLUser + ":" + cfg.MySQLPassword + "@tcp(" + cfg.MySQLHost + ":" + cfg.MySQLPort + ")/" + cfg.MySQLDatabase + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Seed provides some initial data for the database.
func seed(db *gorm.DB) {
	// Get the game with the ID of 1
	var game models.Game
	res := db.First(&game, 1)
	if res.Error != nil {
		log.Println("Creating game")
		// Create a new game
		game = models.Game{
			UUID:    "test",
			Playing: true,
		}
		db.Create(&game)
	}
}

// MigrateDatabase migrates the database.
func MigrateDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Game{}, &models.User{}, &models.Session{})
	seed(db)
	return err
}

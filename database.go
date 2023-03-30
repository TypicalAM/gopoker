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

// delOrphan deletes the orphan games from the database.
func delOrphan(db *gorm.DB) {
	// Set the gameID for every user to 0
	res := db.Model(&models.User{}).Where("game_id IS NOT NULL").Update("game_id", nil)
	log.Println("Cleared games from ", res.RowsAffected, " users")

	// Delete all games
	res = db.Model(&models.Game{}).Where("playing", false).Preload("Players").Delete(&models.Game{})
	log.Println("Deleted ", res.RowsAffected, " orphan games")
}

// MigrateDatabase migrates the database.
func MigrateDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Game{}, &models.User{}, &models.Session{})
	//delOrphan(db)
	return err
}

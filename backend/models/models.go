package models

import (
	"fmt"
	"log"

	"github.com/TypicalAM/gopoker/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New connects to the database using the config.
func New(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Warsaw",
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabasePort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// delOrphan deletes the orphan games from the database.
func delOrphan(db *gorm.DB) {
	// Set the gameID for every user to 0
	res := db.Model(&User{}).Where("game_id IS NOT NULL").Update("game_id", nil)
	log.Println("Cleared games from ", res.RowsAffected, " users")

	// Delete all games
	res = db.Model(&Game{}).Where("playing", false).Preload("Players").Delete(&Game{})
	log.Println("Deleted ", res.RowsAffected, " orphan games")
}

// Migrate migrates the database.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Game{}, &User{}, &Session{}, &Profile{}); err != nil {
		return err
	}

	delOrphan(db)
	return nil
}

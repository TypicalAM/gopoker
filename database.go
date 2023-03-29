package gopoker

import (
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

// MigrateDatabase migrates the database.
func MigrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Session{})
}

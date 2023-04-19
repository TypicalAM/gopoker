package main

import (
	"log"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
)

func main() {
	// Read the config file
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the database
	db, err := models.ConnectToDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the database
	err = models.MigrateDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the router
	router, err := routes.SetupRouter(db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Run the app
	if err = router.Run(cfg.ListenPort); err != nil {
		log.Fatalln(err)
	}
}

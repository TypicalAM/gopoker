package main

import (
	"log"
	"os"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
)

func main() {
	// Get the cwd
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Read the config file
	cfg, err := config.ReadConfig(cwd)
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

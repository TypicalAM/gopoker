package main

import (
	"log"
	"time"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/TypicalAM/gopoker/services"
)

func main() {
	// Read the config file
	cfg := config.New()

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

	// Set up the file service
	fileService, err := services.NewCloudinaryService(cfg.CloudinaryURL, "profile_images", 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the router
	router, err := routes.SetupRouter(db, cfg, fileService)
	if err != nil {
		log.Fatal(err)
	}

	// Run the app
	if err = router.Run(cfg.ListenPort); err != nil {
		log.Fatalln(err)
	}
}

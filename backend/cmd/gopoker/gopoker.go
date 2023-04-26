package main

import (
	"log"
	"time"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/TypicalAM/gopoker/services/upload"
)

func main() {
	// Set up the logger
	log.SetFlags(log.Lshortfile)

	// Read the config file
	cfg := config.New()

	// Connect to the database
	db, err := models.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the database
	err = models.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the file service
	var uploader upload.Uploader
	switch cfg.FileUploadType {
	case config.Local:
		uploader, err = upload.NewLocal(cfg.FileUploadPath)
	case config.Cloudinary:
		uploader, err = upload.NewCloudinary(cfg.CloudinaryURL, "profile_images", 5*time.Second)
	default:
		log.Fatalf("invalid file upload type: %v", cfg.FileUploadType)
	}

	if err != nil {
		log.Fatal(err)
	}

	// Set up the router
	router, err := routes.New(db, cfg, uploader)
	if err != nil {
		log.Fatal(err)
	}

	// Run the app
	if err = router.Run(cfg.ListenPort); err != nil {
		log.Fatalln(err)
	}
}

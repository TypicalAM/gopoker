package main

import (
	"log"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// setupRouter sets up the router
func setupRouter(db *gorm.DB, cfg *config.Config) (*gin.Engine, error) {
	store := cookie.NewStore([]byte(cfg.CookieSecret))

	// Default middleware
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(sessions.Sessions("gopoker_session", store))
	router.Use(middleware.Session(db))
	router.Use(middleware.General())

	// Set up the static file server
	router.StaticFile("/", "./build/index.html")

	// All static assets should be under the /assets path
	assets := router.Group("/assets")
	assets.Use(middleware.Cache(cfg.CacheLifetime))
	assets.Static("/static", "./build/static")

	// Create the controller
	hub := game.NewHub()
	go hub.Run()
	controller := routes.New(db, hub, cfg)

	// Set up the api
	api := router.Group("/api")
	noAuth := api.Group("/")
	noAuth.Use(middleware.NoAuth())
	noAuth.Use(middleware.Throttle(cfg.RequestsPerMin))
	noAuth.POST("/register", controller.Register)
	//noAuth.POST("/login", controller.Login)

	auth := api.Group("/")
	auth.Use(middleware.Auth())
	auth.Use(middleware.Sensitive())
	//auth.GET("/logout", controller.Logout)
	//auth.GET("/game/queue", controller.Queue)
	//auth.GET("/game/id/:id/ws", controller.GameSocket)

	return router, nil
}

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
	router, err := setupRouter(db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Run the app
	if err = router.Run(cfg.ListenPort); err != nil {
		log.Fatalln(err)
	}
}

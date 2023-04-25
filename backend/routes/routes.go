package routes

import (
	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/services/upload"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// controller holds all the variables needed for routes to perform their logic
type controller struct {
	db       *gorm.DB
	hub      *game.Hub
	config   *config.Config
	uploader upload.Uploader
}

// New creates a new router with all the routes
func New(db *gorm.DB, cfg *config.Config, uploader upload.Uploader) (*gin.Engine, error) {
	store := cookie.NewStore([]byte(cfg.CookieSecret))

	// Allow cors
	corsCofig := cors.DefaultConfig()
	corsCofig.AllowOrigins = cfg.CorsTrustedOrigins
	corsCofig.AllowCredentials = true

	// Default middleware
	router := gin.Default()
	router.Use(cors.New(corsCofig))
	router.Use(sessions.Sessions("gopoker_session", store))
	router.Use(middleware.Session(db))
	router.Use(middleware.General())

	// Create the controller
	hub := game.NewHub()
	go hub.Run()
	controller := controller{
		db:       db,
		hub:      hub,
		config:   cfg,
		uploader: uploader,
	}

	// Set up the api
	api := router.Group("/api")
	noAuth := api.Group("/")
	noAuth.Use(middleware.NoAuth())
	noAuth.Use(middleware.Throttle(cfg.RequestsPerMin))
	noAuth.POST("/register", controller.Register)
	noAuth.POST("/login", controller.Login)

	auth := api.Group("/")
	auth.Use(middleware.Auth())
	auth.Use(middleware.Sensitive())
	auth.POST("/logout", controller.Logout)
	auth.POST("/game/queue", controller.Queue)
	auth.GET("/game/id/:id", controller.Game)
	auth.GET("/profile", controller.Profile)
	auth.PUT("/profile", controller.ProfileUpdate)

	return router, nil
}

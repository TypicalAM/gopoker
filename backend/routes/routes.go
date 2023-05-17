package routes

import (
	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/services/game"
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
	gameSrv  *game.Server
	config   *config.Config
	uploader upload.Uploader
}

// New creates a new router with all the routes
func New(db *gorm.DB, cfg *config.Config, uploader upload.Uploader) (*gin.Engine, error) {
	store := cookie.NewStore([]byte(cfg.CookieSecret))

	// Allow cors
	corsCofig := cors.DefaultConfig()
	corsCofig.AllowOrigins = cfg.TrustedOrigins
	corsCofig.AllowCredentials = true

	// Default middleware
	router := gin.Default()
	router.Use(cors.New(corsCofig))
	router.Use(sessions.Sessions("gopoker_session", store))
	router.Use(middleware.Session(db))
	router.Use(middleware.General())

	// Create the game server and the controller
	gameSrv := game.New(db)
	go gameSrv.Run()
	controller := controller{
		db:       db,
		gameSrv:  gameSrv,
		config:   cfg,
		uploader: uploader,
	}

	// Serve the static files if we are uploading to the local file system
	if cfg.FileUploadType == config.Local {
		router.Static("/uploads", cfg.FileUploadPath)
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

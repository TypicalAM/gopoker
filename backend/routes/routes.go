package routes

import (
	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Controller holds all the variables needed for routes to perform their logic
type Controller struct {
	db     *gorm.DB
	hub    *game.Hub
	config *config.Config
}

// New creates a new instance of the Controller
func New(db *gorm.DB, hub *game.Hub, c *config.Config) Controller {
	return Controller{
		db:     db,
		hub:    hub,
		config: c,
	}
}

// SetupRouter sets up the router
func SetupRouter(db *gorm.DB, cfg *config.Config) (*gin.Engine, error) {
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

	// All static assets should be under the /images path
	assets := router.Group("/images")
	assets.Use(middleware.Cache(cfg.CacheLifetime))
	assets.Static("/", "./images/")

	// Create the controller
	hub := game.NewHub()
	go hub.Run()
	controller := New(db, hub, cfg)

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

	return router, nil
}

// isAuthenticated checks if the current user is authenticated or not
func isAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(middleware.UserIDKey)
	return exists
}

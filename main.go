package gopoker

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// staticFS is an embedded file system that contains the static files
//go:embed dist/*
var staticFS embed.FS

// Run is the main function
func Run() {
	// Read the config file
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the database
	db, err := ConnectToDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the database
	err = MigrateDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	// Try to load the templates
	t, err := loadTemplates()
	if err != nil {
		log.Fatal(err)
	}

	// Set up the router
	r := gin.Default()

	// Set up the cookie store
	store := cookie.NewStore([]byte(cfg.CookieSecret))

	// Set up the session middleware
	r.Use(sessions.Sessions("gopoker_session", store))

	// Set up the templates
	r.SetHTMLTemplate(t)

	subFS, err := fs.Sub(staticFS, "dist/assets")
	if err != nil {
		log.Fatal(err)
	}

	// All static assets should be under the /assets path
	assets := r.Group("/assets")
	assets.Use(middleware.Cache(cfg.CacheLifetime))
	assets.StaticFS("/", http.FS(subFS))

	r.Use(middleware.Session(db))
	r.Use(middleware.General())

	controller := routes.New(db, cfg)
	r.Any("/", controller.Index)

	noAuth := r.Group("/")
	noAuth.Use(middleware.NoAuth())
	noAuth.GET("/login", controller.Login)
	noAuth.GET("/register", controller.Register)

	noAuthPost := noAuth.Group("/")
	noAuthPost.Use(middleware.Throttle(cfg.RequestsPerMin))
	noAuthPost.POST("/login", controller.LoginPost)
	noAuthPost.POST("/register", controller.RegisterPost)

	auth := r.Group("/game")
	auth.Use(middleware.Auth())
	auth.Use(middleware.Sensitive())
	auth.GET("/lobby", controller.Lobby)
	auth.GET("/lobby/queue", controller.Queue)
	auth.GET("/id/:id", controller.Game)
	auth.GET("/id/:id/leave", controller.LeaveGame)

	if err = r.Run(cfg.ListenPort); err != nil {
		log.Fatalln(err)
	}
}

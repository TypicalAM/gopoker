package routes

import (
	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/game"
	"github.com/TypicalAM/gopoker/middleware"
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

// isAuthenticated checks if the current user is authenticated or not
func isAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(middleware.UserIDKey)
	return exists
}

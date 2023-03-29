package routes

import (
	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Controller holds all the variables needed for routes to perform their logic
type Controller struct {
	db     *gorm.DB
	config *config.Config
}

// New creates a new instance of the routes.Controller
func New(db *gorm.DB, c *config.Config) Controller {
	return Controller{
		db:     db,
		config: c,
	}
}

// PageData holds the default data needed for HTML pages to render
type PageData struct {
	Title           string
	Messages        []Message
	IsAuthenticated bool
	CacheParameter  string
	Trans           func(s string) string
}

// Message holds a message which can be rendered as responses on HTML pages
type Message struct {
	Type    string // success, warning, error, etc.
	Content string
}

// isAuthenticated checks if the current user is authenticated or not
func isAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(middleware.UserIDKey)
	return exists
}

func (controller Controller) DefaultPageData(c *gin.Context) PageData {
	return PageData{
		Title:           "Home",
		Messages:        nil,
		IsAuthenticated: isAuthenticated(c),
		CacheParameter:  controller.config.CacheParameter,

		// TODO: Dirty hack, fix this
		Trans: func(str string) string { return str },
	}
}

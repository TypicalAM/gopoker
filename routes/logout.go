package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Logout logs the user out
func (controller Controller) Logout(c *gin.Context) {
	// Clear the session
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	// Redirect to the login page
	c.Redirect(http.StatusFound, "/login")
}

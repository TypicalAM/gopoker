package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserIDKey is the key used to set and get the user id in the context of the current request
const UserIDKey = "UserID"

// Auth middleware redirects to /login and aborts the current request if there is no authenticated user
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(UserIDKey)
		if !exists {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}
	}
}

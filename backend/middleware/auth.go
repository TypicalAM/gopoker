package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserIDKey is the key used to set and get the user id in the context of the current request
const UserIDKey = "UserID"

// Auth middleware checks if the user is logged in
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(UserIDKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}
	}
}

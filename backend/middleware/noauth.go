package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NoAuth is for routes that can only be acccessed while not being authenticated
func NoAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(UserIDKey)
		if !exists {
			c.Next()
			return
		}

		c.JSON(http.StatusConflict, gin.H{"error": "You are already authenticated"})
		c.Abort()
	}
}

package routes

import (
	"fmt"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-gonic/gin"
)

// GetUser retrieves the current user from the context
func (conttroller Controller) GetUser(c *gin.Context) (*models.User, error) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		return nil, fmt.Errorf("user id does not exist in context")
	}

	var user models.User
	res := conttroller.db.Model(&models.User{}).Preload("Profile").Where("id = ?", userID).Find(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

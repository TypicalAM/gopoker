package routes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/TypicalAM/gopoker/middleware"
	"github.com/TypicalAM/gopoker/models"
	"github.com/gin-gonic/gin"
)

// imageFromBase64 converts a base64 string to an image
// TODO: This is not the best way to do this, but it works for now
func isImage(data string) bool {
	decoded, err := base64.StdEncoding.DecodeString(data[strings.IndexByte(data, ',')+1:])
	if err != nil || !strings.HasPrefix(http.DetectContentType(decoded), "image/") {
		return false
	}

	return true
}

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

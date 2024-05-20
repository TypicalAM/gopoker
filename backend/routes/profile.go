package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Profile fetches the user's profile.
func (con controller) Profile(c *gin.Context) {
	user, err := con.getUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user.Sanitize()})
}

type ProfileUpdateData struct {
	DisplayName string `json:"display_name,omitempty"`
	ImageData   string `json:"image_data,omitempty"`
}

// ProfileUpdate updates the user's profile.
func (con controller) ProfileUpdate(c *gin.Context) {
	user, err := con.getUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var userUpdateData ProfileUpdateData
	err = c.ShouldBindJSON(&userUpdateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	var displayNameUpdate bool
	if userUpdateData.DisplayName != "" {
		displayNameUpdate = true
		user.Profile.DisplayName = userUpdateData.DisplayName
		if res := con.db.Model(user.Profile).Where("user_id = ?", user.ID).Updates(user.Profile); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving"})
			return
		}
	}

	if userUpdateData.ImageData == "" {
		if displayNameUpdate {
			c.JSON(http.StatusOK, gin.H{"user": user.Sanitize()})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "no data to update"})
	}

	url, err := con.uploader.UploadFile(userUpdateData.ImageData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = con.uploader.DeleteFile(user.Profile.ImageURL); err != nil {
		log.Println("error deleting old image:", err)
	}

	user.Profile.ImageURL = url
	if res := con.db.Model(user.Profile).Where("user_id = ?", user.ID).Updates(user.Profile); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user.Sanitize()})
}

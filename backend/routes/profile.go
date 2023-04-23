package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

// Profile fetches the user's profile.
func (controller Controller) Profile(c *gin.Context) {
	user, err := controller.GetUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

type ProfileUpdateData struct {
	DisplayName string `json:"display_name,omitempty"`
	ImageData   string `json:"image_data,omitempty"`
}

// ProfileUpdate updates the user's profile.
func (controller Controller) ProfileUpdate(c *gin.Context) {
	user, err := controller.GetUser(c)
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
		if res := controller.db.Save(&user); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving"})
			return
		}
	}

	if userUpdateData.ImageData == "" {
		if displayNameUpdate {
			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "no data to update"})
	}

	if ok := isImage(userUpdateData.ImageData); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}

	cld, err := cloudinary.NewFromURL(controller.config.CloudinaryURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error uploading image"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uploadResult, err := cld.Upload.Upload(
		ctx,
		userUpdateData.ImageData,
		uploader.UploadParams{
			Folder: "profile_images",
		})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error uploading image"})
		return
	}

	user.Profile.ImageURL = uploadResult.SecureURL
	res := controller.db.Model(user.Profile).Where("user_id = ?", user.ID).Updates(user.Profile)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

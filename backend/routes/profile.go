package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	Password    string `json:"password,omitempty"`
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

	log.Printf("%+v\n", userUpdateData)

	if userUpdateData.DisplayName != "" {
		user.Profile.DisplayName = userUpdateData.DisplayName
		if res := con.db.Model(user.Profile).Where("user_id = ?", user.ID).Updates(user.Profile); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving"})
			return
		}
	}

	if userUpdateData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userUpdateData.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}

		user.Password = string(hashedPassword)
		if res := con.db.Save(&user); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user"})
			return
		}
	}

	if userUpdateData.ImageData != "" {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"user": user.Sanitize()})
}

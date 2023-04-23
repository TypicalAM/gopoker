package services

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Imageinterface for file upload services.
type Image interface {
	UploadFile(data string) (string, error)
	DeleteFile(url string) error
}

var ErrInvalidImage = errors.New("invalid image")

// CloudinaryService is a service for uploading files to Cloudinary.
type CloudinaryService struct {
	cld        *cloudinary.Cloudinary
	folderName string
	timeout    time.Duration
}

// NewCloudinaryService creates a new CloudinaryService.
func NewCloudinaryService(url string, folderName string, timeout time.Duration) (Image, error) {
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		return CloudinaryService{}, err
	}

	return CloudinaryService{
		cld:        cld,
		folderName: folderName,
		timeout:    timeout,
	}, nil
}

// UploadFile uploads a file to Cloudinary.
func (service CloudinaryService) UploadFile(data string) (string, error) {
	if !isImage(data) {
		return "", ErrInvalidImage
	}

	ctx, cancel := context.WithTimeout(context.Background(), service.timeout)
	defer cancel()

	res, err := service.cld.Upload.Upload(ctx, data, uploader.UploadParams{Folder: service.folderName})
	if err != nil {
		return "", err
	}

	return res.SecureURL, nil
}

// DeleteFile deletes a file from Cloudinary.
func (service CloudinaryService) DeleteFile(url string) error {
	// TODO: Implement
	return nil
}

// isImage checks if the data is an image.
// TODO: This is not the best way to do this, but it works for now
func isImage(data string) bool {
	decoded, err := base64.StdEncoding.DecodeString(data[strings.IndexByte(data, ',')+1:])
	if err != nil || !strings.HasPrefix(http.DetectContentType(decoded), "image/") {
		return false
	}

	return true
}

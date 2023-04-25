package upload

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryService is a service for uploading files to Cloudinary.
type CloudinaryService struct {
	cld        *cloudinary.Cloudinary
	folderName string
	timeout    time.Duration
}

// NewCloudinary creates a new CloudinaryService.
func NewCloudinary(url string, folderName string, timeout time.Duration) (Uploader, error) {
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

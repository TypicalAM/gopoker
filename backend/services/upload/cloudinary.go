package upload

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryUploader is a service for uploading files to Cloudinary.
type CloudinaryUploader struct {
	cld        *cloudinary.Cloudinary
	folderName string
	timeout    time.Duration
}

// NewCloudinary creates a new CloudinaryService.
func NewCloudinary(url string, folderName string, timeout time.Duration) (Uploader, error) {
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		return CloudinaryUploader{}, err
	}

	return CloudinaryUploader{
		cld:        cld,
		folderName: folderName,
		timeout:    timeout,
	}, nil
}

// UploadFile uploads a file to Cloudinary.
func (u CloudinaryUploader) UploadFile(data string) (string, error) {
	if !isImage(data) {
		return "", ErrInvalidImage
	}

	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	res, err := u.cld.Upload.Upload(ctx, data, uploader.UploadParams{Folder: u.folderName})
	if err != nil {
		return "", err
	}

	return res.SecureURL, nil
}

// DeleteFile deletes a file from Cloudinary.
func (u CloudinaryUploader) DeleteFile(url string) error {
	// TODO: Implement
	return nil
}

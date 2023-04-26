package upload

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalUploader is a service for uploading files to the local file system
type LocalUploader struct {
	path string
}

// NewLocal creates a new local file uploader
func NewLocal(path string) (LocalUploader, error) {
	return LocalUploader{
		path: filepath.Clean(path),
	}, nil
}

// UploadFile uploads a file to the local file system
func (u LocalUploader) UploadFile(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data[strings.IndexByte(data, ',')+1:])
	if err != nil {
		return "", ErrInvalidImage
	}

	var ext string
	mimeType := http.DetectContentType(decoded)
	switch mimeType {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	default:
		log.Println("invalid mime type:", mimeType)
		return "", ErrInvalidImage
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	if err = os.WriteFile(filepath.Join(u.path, filename), decoded, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("/uploads/%s", filename), nil
}

// DeleteFile deletes a file from the local file system
func (u LocalUploader) DeleteFile(url string) error {
	if url == "" {
		return ErrNonExistentFile
	}

	split := strings.Split(url, "/")
	filename := filepath.Clean(split[len(split)-1])
	if _, err := os.Stat(filepath.Join(u.path, filename)); err != nil {
		return ErrNonExistentFile
	}

	if err := os.Remove(filepath.Join(u.path, filename)); err != nil {
		return err
	}

	return nil
}

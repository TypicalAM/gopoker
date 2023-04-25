package upload

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// ErrInvalidImage is returned when the data is not an image.
var ErrInvalidImage = errors.New("invalid image")

// Uploader interface for file upload services.
type Uploader interface {
	UploadFile(data string) (string, error)
	DeleteFile(url string) error
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

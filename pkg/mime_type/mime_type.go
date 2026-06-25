package mimetype

import "strings"

var mimeToPreviewType = map[string]string{
	"image/":       "image",
	"video/":       "video",
	"audio/":       "audio",
	"text/":        "site",
	"application/": "file",
}

func GetMimeType(contentType string) string {
	var fileType string
	for prefix, pType := range mimeToPreviewType {
		if strings.HasPrefix(contentType, prefix) {
			fileType = pType
			break
		}
	}

	return fileType
}

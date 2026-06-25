package s3urlparse

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseS3Path(path string) (bucket, object string, err error) {
	// Если передан полный URL
	if strings.Contains(path, "://") {
		u, err := url.Parse(path)
		if err != nil {
			return "", "", err
		}
		path = strings.TrimPrefix(u.Path, "/")
	}

	// Разбиваем на bucket/object
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid format: expected bucket/object, got %s", path)
	}

	return parts[0], parts[1], nil
}

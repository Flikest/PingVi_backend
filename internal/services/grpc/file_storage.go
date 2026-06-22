package servicegrpc

import (
	"context"
	"io"
)

type FileInfo struct {
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified string            `json:"last_modified"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata"`
	IsPublic     bool              `json:"is_public"`
}

type UploadOptions struct {
	ContentType string
	Metadata    map[string]string
	IsPublic    bool
	ChunkSize   int64
}

type DownloadOptions struct {
	VersionID string
}

type ListOptions struct {
	Prefix    string
	Recursive bool
	MaxKeys   int
}

type FileStorageRepository interface {
	UploadFileChunked(ctx context.Context, reader io.Reader, fileName string, opts UploadOptions) (*FileInfo, error)
}

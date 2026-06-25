package repository

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

func (r *RepositoryFileStorage) UploadFile(ctx context.Context, bucketName string, objectName string, file io.Reader, size int64) (string, error) {
	info, err := r.MinIOClient.PutObject(ctx, bucketName, objectName, file, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		r.Log.Error("error with uploading file to s3: ", "error", err)
		return "", err
	}

	return info.Location, nil
}

func (r *RepositoryFileStorage) GenerateTemporaryURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration, reqParams url.Values) (string, error) {
	presignedURL, err := r.MinIOClient.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return presignedURL.String(), nil
}

func (r *RepositoryFileStorage) DeleteFile(ctx context.Context, bucketName string, objectName string) error {
	err := r.MinIOClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		r.Log.Error("error with deleting file from s3: ", "error", err)
		return err
	}

	return nil
}

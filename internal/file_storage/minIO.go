package filestorage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	BucketPrivate = "private"
	BucketPublic  = "public"
)

type MinIOConfig struct {
	Endpoint string
	UseSSL   bool
	Logger   *slog.Logger
}

func NewMinIOClient(ctx context.Context, conf MinIOConfig) (*minio.Client, error) {
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")

	if accessKey == "" || secretKey == "" {
		return nil, errors.New("MINIO_ROOT_USER and MINIO_ROOT_PASSWORD must be set")
	}

	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: conf.UseSSL,
	})
	if err != nil {
		conf.Logger.Error("failed to create MinIO client", "error", err)
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	conf.Logger.Info("MinIO client created successfully", "endpoint", conf.Endpoint)

	if err := ensureBuckets(ctx, client, conf.Logger); err != nil {
		return nil, err
	}

	if err := setupPublicBucket(ctx, client, conf.Logger); err != nil {
		return nil, err
	}

	return client, nil
}

func ensureBuckets(ctx context.Context, client *minio.Client, logger *slog.Logger) error {
	buckets := []string{BucketPrivate, BucketPublic}

	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			logger.Error("failed to check bucket existence", "bucket", bucket, "error", err)
			return fmt.Errorf("check bucket %s: %w", bucket, err)
		}

		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
				Region: "us-east-1",
			})
			if err != nil {
				logger.Error("failed to create bucket", "bucket", bucket, "error", err)
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
			logger.Info("Bucket created successfully", "bucket", bucket)
		} else {
			logger.Debug("Bucket already exists", "bucket", bucket)
		}
	}

	return nil
}

func setupPublicBucket(ctx context.Context, client *minio.Client, logger *slog.Logger) error {
	publicPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::` + BucketPublic + `/*"]
			}
		]
	}`

	err := client.SetBucketPolicy(ctx, BucketPublic, publicPolicy)
	if err != nil {
		logger.Error("failed to set public bucket policy", "bucket", BucketPublic, "error", err)
		return fmt.Errorf("set bucket policy: %w", err)
	}

	logger.Info("Public bucket configured successfully", "bucket", BucketPublic)
	return nil
}

func GetBucketPolicy(ctx context.Context, client *minio.Client, bucketName string) (string, error) {
	return client.GetBucketPolicy(ctx, bucketName)
}

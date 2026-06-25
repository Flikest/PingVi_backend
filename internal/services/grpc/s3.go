package servicegrpc

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	s3urlparse "github.com/Flikest/PingVi_backend/pkg/s3_url_parse"
)

type RepositoryS3 interface {
	GenerateTemporaryURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration, reqParams url.Values) (string, error)
}

func (s *ServiceS3) GenerateTemporaryURL(ctx context.Context, links []string) ([]dto.IssueTemporaryURLResponse, error) {
	if len(links) == 0 {
		s.Log.Warn("no links provided")
		return nil, errors.New("no links provided")
	}

	var result []dto.IssueTemporaryURLResponse

	for _, link := range links {
		if strings.HasPrefix(link, "public/") {
			result = append(result, dto.IssueTemporaryURLResponse{
				TemporaryURL: link,
			})
			continue
		}

		// http://localhost:9000/private/catalog/uuid.ext
		bucketName, objectName, err := s3urlparse.ParseS3Path(link)
		if err != nil {
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "invalid filepath: " + err.Error(),
			})
			continue
		}

		if bucketName != "private" {
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "temporary URLs can only be generated for private files",
			})
			continue
		}

		temporaryURL, err := s.Repository.GenerateTemporaryURL(
			ctx,
			bucketName,
			objectName,
			15*time.Minute,
			nil,
		)

		if err != nil {
			s.Log.Error("failed to generate temporary URL",
				"bucket", bucketName,
				"object", objectName,
				"error", err)
			result = append(result, dto.IssueTemporaryURLResponse{
				Error: "failed to generate temporary URL: " + err.Error(),
			})
			continue
		}

		result = append(result, dto.IssueTemporaryURLResponse{
			TemporaryURL: temporaryURL,
		})
	}

	return result, nil
}

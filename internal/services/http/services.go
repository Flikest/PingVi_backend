package servicehttp

import (
	"context"
	"log/slog"

	pbUserPermissions "github.com/Flikest/PingVi_backend/gen/go/permissions"
	pbUserInfo "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/repository"
	"github.com/minio/minio-go/v7"
)

type ServiceSSO struct {
	Log        *slog.Logger
	Repository *repository.RepositorySSO
}

type ServiceMessenger struct {
	Log        *slog.Logger
	Hub        *Hub
	Client     pbUserInfo.UserInfoClient
	Ctx        context.Context
	Repository *repository.RepositoryMessenger
}

type ServiceSFU struct {
	Log        *slog.Logger
	GrpcClient pbUserPermissions.PermissionsClient
}

type ServiceS3 struct {
	Log         *slog.Logger
	MinIOClient *minio.Client
}

func NewSSOService(s *ServiceSSO) *ServiceSSO {
	return &ServiceSSO{
		Log:        s.Log,
		Repository: s.Repository,
	}
}

func NewMessengerService(s *ServiceMessenger) *ServiceMessenger {
	return &ServiceMessenger{
		Log:        s.Log,
		Hub:        s.Hub,
		Repository: s.Repository,
	}
}

func NewSFUService(s *ServiceSFU) *ServiceSFU {
	return &ServiceSFU{
		Log:        s.Log,
		GrpcClient: s.GrpcClient,
	}
}

func NewS3Service(s *ServiceS3) *ServiceS3 {
	return &ServiceS3{
		Log:         s.Log,
		MinIOClient: s.MinIOClient,
	}
}

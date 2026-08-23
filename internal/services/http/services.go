package servicehttp

import (
	"log/slog"

	pbUserPermissions "github.com/Flikest/PingVi_backend/gen/go/permissions"
	pbUserInfo "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/repository"
)

type ServiceSSO struct {
	Log        *slog.Logger
	Repository *repository.RepositorySSO
}

type ServiceMessenger struct {
	Log        *slog.Logger
	Hub        *Hub
	Client     pbUserInfo.UserInfoClient
	Repository *repository.RepositoryMessenger
}

type ServiceSFU struct {
	Log        *slog.Logger
	GrpcClient pbUserPermissions.PermissionsClient
}

type ServiceFileStorage struct {
	Log        *slog.Logger
	Repository *repository.RepositoryFileStorage
}

type ServiceBots struct {
	BotHub     BotHub
	Log        *slog.Logger
	Repository *repository.RepositoryBots
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

func NewFileStorageService(s *ServiceFileStorage) *ServiceFileStorage {
	return &ServiceFileStorage{
		Log:        s.Log,
		Repository: s.Repository,
	}
}

func NewBotsService(s *ServiceBots) *ServiceBots {
	return &ServiceBots{
		BotHub:     s.BotHub,
		Log:        s.Log,
		Repository: s.Repository,
	}
}

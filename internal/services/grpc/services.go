package servicegrpc

import (
	"log/slog"
)

type ServiceUserInfo struct {
	Log        *slog.Logger
	Repository UserInfoRepository
}

type ServiceS3 struct {
	Log        *slog.Logger
	Repository RepositoryS3
}

type ServicePermissions struct {
	Log        *slog.Logger
	Repository PermissionsRepository
}

type ServiceBots struct {
	Log        *slog.Logger
	Repository BotsRepository
}

func NewServiceUserInfo(log *slog.Logger, repository UserInfoRepository) *ServiceUserInfo {
	return &ServiceUserInfo{
		Log:        log,
		Repository: repository,
	}
}

func NewServiceS3(log *slog.Logger, repository RepositoryS3) *ServiceS3 {
	return &ServiceS3{
		Log:        log,
		Repository: repository,
	}
}

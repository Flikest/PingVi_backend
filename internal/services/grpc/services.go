package servicegrpc

import (
	"log/slog"
)

type ServiceUserInfo struct {
	Log        *slog.Logger
	Repository UserInfoRepository
}

type FileStorage struct {
	Log        *slog.Logger
	Repository FileStorageRepository
}

type ServicePermissions struct {
	Log        *slog.Logger
	Repository PermissionsRepository
}

func NewServiceUserInfo(log *slog.Logger, repository UserInfoRepository) *ServiceUserInfo {
	return &ServiceUserInfo{
		Log:        log,
		Repository: repository,
	}
}

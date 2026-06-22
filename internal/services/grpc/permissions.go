package servicegrpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type PermissionsRepository interface {
	GetUserPermissions(ctx context.Context, userID uuid.UUID) (string, error)
}

func (s *ServicePermissions) GetUserPermission(ctx context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		s.Log.Error("invalid user id")
		return "", errors.New("invalid user id")
	}

	permission, err := s.Repository.GetUserPermissions(ctx, userID)
	if err != nil {
		s.Log.Error("error with getting permission: ", "error", err)
		return "", err
	}

	return permission, nil
}

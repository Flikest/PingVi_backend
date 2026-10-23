package servicegrpc

import (
	"context"
	"errors"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

type UserInfoRepository interface {
	SelectUserNameByID(ctx context.Context, userID uuid.UUID) (name string, err error)
	SelectUserIDBySessionID(ctx context.Context, sessionID string) (userID uuid.UUID, err error)
	SelectUsersByIDs(ctx context.Context, userID []uuid.UUID) (user []dto.User, err error)
}

func (s *ServiceUserInfo) GetUserNameByID(ctx context.Context, userID uuid.UUID) (name string, err error) {
	if userID == uuid.Nil {
		s.Log.Error("invalid user id")
		return "", errors.New("invalid user id")
	}

	name, err = s.Repository.SelectUserNameByID(ctx, userID)
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		return "", err
	}

	return name, nil
}

func (s *ServiceUserInfo) GetUserIDBySessionID(ctx context.Context, sessionID string) (userID uuid.UUID, err error) {
	if sessionID == "" {
		s.Log.Error("empty session id")
		return uuid.Nil, errors.New("empty session id")
	}

	userID, err = s.Repository.SelectUserIDBySessionID(ctx, sessionID)
	if err != nil {
		s.Log.Error("error with getting user id: ", "error", err)
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *ServiceUserInfo) GetUsersByIds(ctx context.Context, userIDs []uuid.UUID) (user []dto.User, err error) {
	if len(userIDs) <= 0 {
		return nil, errors.New("invalid user id")
	}

	user, err = s.Repository.SelectUsersByIDs(ctx, userIDs)
	if err != nil {
		s.Log.Error("error with selecting user by id: ", "error", err)
		return nil, err
	}

	return user, nil
}

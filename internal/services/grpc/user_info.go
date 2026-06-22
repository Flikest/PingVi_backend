package servicegrpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type UserInfoRepository interface {
	SelectUserNameByID(ctx context.Context, userID uuid.UUID) (name string, err error)
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

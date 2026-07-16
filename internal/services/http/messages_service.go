package servicehttp

import (
	"context"
	"errors"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) GetSubscribers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID

	channelIDs, err := s.Repository.SelectChannelIdsByUserId(ctx, userID)
	if err != nil {
		s.Log.Error("error with getting channel ids: ", "error", err)
		return nil, err
	}

	for _, id := range channelIDs {
		userIDs, err := s.Repository.SelectChannelMembers(ctx, id)
		if err != nil {
			s.Log.Error("error with getting members in channel by channel id: ", "error", err)
			return nil, err
		}
		userIDs = append(userIDs, userIDs...)
	}

	groupIDs, err := s.Repository.SelectGroupIdsByUserId(ctx, userID)
	if err != nil {
		s.Log.Error("error with getting group ids: ", "error", err)
		return nil, err
	}

	for _, id := range groupIDs {
		userIDs, err := s.Repository.SelectGroupMembers(ctx, id)
		if err != nil {
			s.Log.Error("error with getting members in group by group id: ", "error", err)
			return nil, err
		}
		userIDs = append(userIDs, userIDs...)
	}

	personalChatIDs, err := s.Repository.SelectPersonalChatMemberIdsByUserId(ctx, userID)
	if err != nil {
		s.Log.Error("error with getting personal chat member ids by user id: ", "error", err)
		return nil, err
	}
	userIDs = append(userIDs, personalChatIDs...)

	return userIDs, nil
}

func (s *ServiceMessenger) ClearMessagesFromCommunity(ctx context.Context, clearMessage dto.ClearMessage) (uuid.UUID, error) {
	switch clearMessage.CommunityType {
	case "group":
		isAdmin, err := s.Repository.SelectGroupMemberIsAdmin(ctx, clearMessage.CommunityID, clearMessage.UserID)
		if err != nil {
			s.Log.Error("error with selecting rights: ", "error", err)
			return uuid.Nil, err
		}

		ownerID, err := s.Repository.SelectGroupOwnerID(ctx, clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting owner: ", "error", err)
			return uuid.Nil, err
		}

		if !isAdmin && ownerID != clearMessage.UserID {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		if err := s.Repository.DeleteMessagesByChatID(ctx, clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from group: ", "error", err)
			return uuid.Nil, err
		}

	case "channel":
		permissions, err := s.Repository.SelectPermissions(ctx, clearMessage.UserID, clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting permissions from channel: ", "error", err)
			return uuid.Nil, err
		}

		if len(permissions) < 3 || permissions[2] != '1' {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		topics, err := s.Repository.SelectTopicsByChannelID(ctx, clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting topics from channel: ", "error", err)
			return uuid.Nil, err
		}

		if len(topics) == 0 {
			s.Log.Warn("no topic found")
			return uuid.Nil, errors.New("no topic found")
		}

		if err := s.Repository.DeleteMessagesByChatID(ctx, clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from topic: ", "error", err)
			return uuid.Nil, err
		}

	case "personal_chat":
		user1ID, user2ID, err := s.Repository.SelectPersonalChatUsers(ctx, clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting user1 and user2 from personal chat", "error", err)
			return uuid.Nil, err
		}

		if clearMessage.UserID != user1ID && clearMessage.UserID != user2ID {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		if err := s.Repository.DeleteMessagesByChatID(ctx, clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from topic: ", "error", err)
			return uuid.Nil, err
		}

	default:
		s.Log.Error("invalid community type")
		return uuid.Nil, errors.New("invalid community type")
	}

	return clearMessage.CommunityID, nil
}

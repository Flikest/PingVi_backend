package servicehttp

import (
	"errors"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) AddMessage(ctx *gin.Context, message dto.AddMessage) (dto.Message, error) {
	// Бизнес-логика: генерация ID и времени
	messageID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating message id: ", "error", err)
		return dto.Message{}, err
	}

	now := time.Now()

	// Создание сообщения
	msg := dto.Message{
		ID:          messageID,
		ChatID:      message.ChatID,
		SenderID:    message.SenderID,
		Message:     message.Message,
		MessageType: message.MessageType,
		Attachments: message.Attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Вставка в БД
	if err := s.Repository.InsertMessage(ctx.Request.Context(), msg); err != nil {
		s.Log.Error("error with inserting message: ", "error", err)
		return dto.Message{}, err
	}

	return msg, nil
}

func (s *ServiceMessenger) UpdateMessage(ctx *gin.Context, message dto.UpdateMessage) (dto.Message, error) {
	// Проверка, что отправитель обновляет свое сообщение
	// В реальном приложении нужно проверить права

	now := time.Now()

	// Обновление сообщения
	msg := dto.Message{
		ID:          message.ID,
		ChatID:      message.ChatID,
		SenderID:    message.SenderID,
		Message:     message.Message,
		MessageType: message.MessageType,
		IsEdited:    true,
		ReplyToID:   message.ReplyToID,
		Attachments: message.Attachments,
		Reactions:   nil,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   now,
	}

	if err := s.Repository.UpdateMessage(ctx.Request.Context(), msg); err != nil {
		s.Log.Error("error with updating message: ", "error", err)
		return dto.Message{}, err
	}

	return msg, nil
}

func (s *ServiceMessenger) DeleteMessage(ctx *gin.Context, message dto.DeleteMessage) error {
	// Проверка, что отправитель удаляет свое сообщение
	// В реальном приложении нужно проверить права

	if err := s.Repository.DeleteMessage(ctx.Request.Context(), message.ID, message.ChatID, message.SenderID); err != nil {
		s.Log.Error("error with deleting message: ", "error", err)
		return err
	}

	return nil
}

func (s *ServiceMessenger) ClearMessagesFromCommunity(ctx *gin.Context, clearMessage dto.ClearMessage) (uuid.UUID, error) {
	switch clearMessage.CommunityType {
	case "group":
		// Бизнес-логика: проверка прав для группы
		isAdmin, err := s.Repository.SelectGroupMemberIsAdmin(ctx.Request.Context(), clearMessage.CommunityID, clearMessage.UserID)
		if err != nil {
			s.Log.Error("error with selecting rights: ", "error", err)
			return uuid.Nil, err
		}

		ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting owner: ", "error", err)
			return uuid.Nil, err
		}

		if !isAdmin && ownerID != clearMessage.UserID {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		// Удаление сообщений
		if err := s.Repository.DeleteMessagesByChatID(ctx.Request.Context(), clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from group: ", "error", err)
			return uuid.Nil, err
		}

	case "channel":
		// Бизнес-логика: проверка прав для канала
		permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), clearMessage.UserID, clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting permissions from channel: ", "error", err)
			return uuid.Nil, err
		}

		if len(permissions) < 3 || permissions[2] != '1' {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		// Проверка тем в канале
		topics, err := s.Repository.SelectTopicsByChannelID(ctx.Request.Context(), clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting topics from channel: ", "error", err)
			return uuid.Nil, err
		}

		if len(topics) == 0 {
			s.Log.Warn("no topic found")
			return uuid.Nil, errors.New("no topic found")
		}

		// Удаление сообщений
		if err := s.Repository.DeleteMessagesByChatID(ctx.Request.Context(), clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from topic: ", "error", err)
			return uuid.Nil, err
		}

	case "personal_chat":
		// Бизнес-логика: проверка прав для личного чата
		user1ID, user2ID, err := s.Repository.SelectPersonalChatUsers(ctx.Request.Context(), clearMessage.CommunityID)
		if err != nil {
			s.Log.Error("error with selecting user1 and user2 from personal chat", "error", err)
			return uuid.Nil, err
		}

		if clearMessage.UserID != user1ID && clearMessage.UserID != user2ID {
			s.Log.Warn("not enough rights")
			return uuid.Nil, errors.New("not enough rights")
		}

		// Удаление сообщений
		if err := s.Repository.DeleteMessagesByChatID(ctx.Request.Context(), clearMessage.CommunityID); err != nil {
			s.Log.Error("error with clearing messages from topic: ", "error", err)
			return uuid.Nil, err
		}

	default:
		s.Log.Error("invalid community type")
		return uuid.Nil, errors.New("invalid community type")
	}

	return clearMessage.CommunityID, nil
}

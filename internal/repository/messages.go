package repository

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) DeleteMessagesByChatID(ctx context.Context, chatID uuid.UUID) error {
	query := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.messages WHERE chat_id=?`,
		[]string{":chat_id"}).
		BindMap(map[string]interface{}{
			":chat_id": chatID,
		})

	if err := query.ExecRelease(); err != nil {
		r.Log.Error("error with clearing messages from chat: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) InsertMessage(ctx context.Context, message dto.Message) error {
	addMessage := r.Session.ContextQuery(
		ctx,
		`INSERT INTO messenger_keyspace.messages (id, chat_id, sender_id, message, message_type, attachments, created_at)
		VALUES (?,?,?,?,?,?,?)`,
		[]string{
			":id",
			":chat_id",
			":sender_id",
			":message",
			":message_type",
			":attachments",
			":created_at"}).
		BindMap(map[string]interface{}{
			":id":           message.ID,
			":chat_id":      message.ChatID,
			":sender_id":    message.SenderID,
			":message":      message.Message,
			":message_type": message.MessageType,
			":attachments":  message.Attachments,
			":created_at":   message.CreatedAt,
		})

	if err := addMessage.ExecRelease(); err != nil {
		r.Log.Error("error with inserting message: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectMessagesByChatID(ctx context.Context, chatID uuid.UUID) ([]dto.Message, error) {
	q := r.Session.ContextQuery(ctx, `SELECT * FROM messenger_keyspace.messages WHERE chat_id=?`, []string{":chat_id"}).
		BindMap(map[string]interface{}{
			":chat_id": chatID,
		})

	var messages []dto.Message
	if err := q.SelectRelease(&messages); err != nil {
		r.Log.Error("error with selecting all messages from chat: ", "error", err)
		return nil, err
	}

	return messages, nil
}

func (r *RepositoryMessenger) UpdateMessage(ctx context.Context, message dto.Message) error {
	updateMessage := r.Session.ContextQuery(
		ctx,
		`UPDATE messenger_keyspace.messages
		SET message=?, is_edited=?, attachments=?, updated_at=?
		WHERE id=? AND chat_id=? AND sender_id=?`,
		[]string{
			":message",
			":is_edited",
			":attachments",
			":updated_at",
			":id",
			":chat_id",
			":sender_id",
		}).BindMap(map[string]interface{}{
		":message":     message.Message,
		":is_edited":   message.IsEdited,
		":attachments": message.Attachments,
		":updated_at":  message.UpdatedAt,
		":id":          message.ID,
		":chat_id":     message.ChatID,
		":sender_id":   message.SenderID,
	})

	if err := updateMessage.ExecRelease(); err != nil {
		r.Log.Error("error with updating message: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectTopicsByChannelID(ctx context.Context, channelID uuid.UUID) ([]dto.Topic, error) {
	selectTopics := r.Session.ContextQuery(
		ctx,
		`SELECT id FROM messenger_keyspace.topics WHERE channel_id=?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	var topicsIDs []dto.Topic
	if err := selectTopics.SelectRelease(&topicsIDs); err != nil {
		r.Log.Error("error with selecting topics from channel: ", "error", err)
		return nil, err
	}

	return topicsIDs, nil
}

func (r *RepositoryMessenger) SelectPersonalChatUsers(ctx context.Context, chatID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	selectUsersFromPersonalChat := r.Session.ContextQuery(
		ctx,
		`SELECT user1_id, user2_id FROM messenger_keyspace.personal_chats WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": chatID,
		})

	var users struct{ user1ID, user2ID uuid.UUID }
	if err := selectUsersFromPersonalChat.SelectRelease(&users); err != nil {
		r.Log.Error("error with selecting user1 and user2 from personal chat", "error", err)
		return uuid.Nil, uuid.Nil, err
	}

	return users.user1ID, users.user2ID, nil
}

func (r *RepositoryMessenger) DeleteMessage(ctx context.Context, messageID, chatID, senderID uuid.UUID) error {
	deleteMessage := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.messages WHERE id=? AND chat_id=? AND sender_id=?`,
		[]string{
			":id",
			":chat_id",
			":sender_id",
		}).BindMap(map[string]interface{}{
		":id":        messageID,
		":chat_id":   chatID,
		":sender_id": senderID,
	})

	if err := deleteMessage.ExecRelease(); err != nil {
		r.Log.Error("error with deleting message: ", "error", err)
		return err
	}

	return nil
}

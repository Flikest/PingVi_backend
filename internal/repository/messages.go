package repository

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) SelectUnreadCountMessage(ctx context.Context, userID, chatID uuid.UUID) (int64, error) {
	var lastReadID snowflake.ID
	query := r.Session.ContextQuery(
		ctx,
		`SELECT last_read_message_id FROM messenger_keyspace.user_read_positions WHERE user_id=? AND chat_id=?`,
		[]string{":user_id", ":chat_id"},
	).BindMap(map[string]interface{}{
		":user_id": userID,
		":chat_id": chatID,
	})

	var result struct {
		LastReadMessageID snowflake.ID `db:"last_read_message_id"`
	}

	if err := query.SelectRelease(&result); err != nil {
		r.Log.Warn("no read position found for user, counting all messages as unread", "user_id", userID, "chat_id", chatID)
		lastReadID = 0
	} else {
		lastReadID = result.LastReadMessageID
	}

	var buckets []int64
	bucketQuery := r.Session.ContextQuery(
		ctx,
		`SELECT DISTINCT backet_id FROM messenger_keyspace.messages WHERE chat_id=?`,
		[]string{":chat_id"},
	).BindMap(map[string]interface{}{
		":chat_id": chatID,
	})

	if err := bucketQuery.SelectRelease(&buckets); err != nil {
		r.Log.Error("error getting distinct backet_ids", "error", err)
		return 0, err
	}

	if len(buckets) == 0 {
		return 0, nil
	}

	var totalUnread int64 = 0

	for _, bucketID := range buckets {
		countQuery := r.Session.ContextQuery(
			ctx,
			`SELECT COUNT(*) FROM messenger_keyspace.messages 
             WHERE chat_id=? AND backet_id=? AND id>?`,
			[]string{":chat_id", ":backet_id", ":last_read_id"},
		).BindMap(map[string]interface{}{
			":chat_id":      chatID,
			":backet_id":    bucketID,
			":last_read_id": lastReadID,
		})

		var count int64
		if err := countQuery.SelectRelease(&count); err != nil {
			r.Log.Error("error counting unread messages in bucket", "error", err, "bucket", bucketID)
			return 0, err
		}
		totalUnread += count
	}

	return totalUnread, nil
}

func (r *RepositoryMessenger) SelectLastMessageByChatID(ctx context.Context, chatID uuid.UUID) (*dto.Message, error) {
	var buckets []int64
	bucketQuery := r.Session.ContextQuery(
		ctx,
		`SELECT DISTINCT backet_id FROM messenger_keyspace.messages WHERE chat_id=?`,
		[]string{":chat_id"},
	).BindMap(map[string]interface{}{
		":chat_id": chatID,
	})

	if err := bucketQuery.SelectRelease(&buckets); err != nil {
		r.Log.Error("error getting distinct backet_ids", "error", err)
		return nil, err
	}

	if len(buckets) == 0 {
		return nil, nil
	}

	var lastMessage *dto.Message
	var maxID snowflake.ID = 0

	for _, bucketID := range buckets {
		var messages []dto.Message
		msgQuery := r.Session.ContextQuery(
			ctx,
			`SELECT * FROM messenger_keyspace.messages 
             WHERE chat_id=? AND backet_id=? 
             ORDER BY id DESC LIMIT 1`,
			[]string{":chat_id", ":backet_id"},
		).BindMap(map[string]interface{}{
			":chat_id":   chatID,
			":backet_id": bucketID,
		})

		if err := msgQuery.SelectRelease(&messages); err != nil {
			r.Log.Error("error getting last message from bucket", "error", err, "bucket", bucketID)
			return nil, err
		}

		if len(messages) > 0 && messages[0].ID > maxID {
			maxID = messages[0].ID
			lastMessage = &messages[0]
		}
	}

	return lastMessage, nil
}

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

func (r *RepositoryMessenger) ReadMessage(ctx context.Context, rm dto.ReadMessage) error {
	readMesage := r.Session.ContextQuery(
		ctx,
		`INSERT INTO messenger_keyspace.user_read_positions (user_id, chat_id, last_read_message_id, last_read_time)
		VALUES (?,?,?,?)`,
		[]string{
			":user_id",
			":chat_id",
			":last_read_message_id",
			":last_read_time",
		}).BindMap(
		map[string]interface{}{
			":chat_id":              rm.ChatID,
			":last_read_message_id": rm.MessageID,
			":last_read_time":       rm.At,
		})

	if err := readMesage.ExecRelease(); err != nil {
		r.Log.Error("error with inserting last read mesage: ", "error", err)
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
			":backet_id",
			":sender_id",
			":message",
			":reply_to_id",
			":created_at"}).
		BindMap(map[string]interface{}{
			":id":          message.ID,
			":chat_id":     message.ChatID,
			":backet_id":   message.ID / 10000,
			":sender_id":   message.SenderID,
			":message":     message.Message,
			":reply_to_id": message.ReplyToID,
			":created_at":  message.CreatedAt,
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
		SET message=?, updated_at=?
		WHERE id=? AND chat_id=? AND sender_id=?`,
		[]string{
			":message",
			":updated_at",
			":id",
			":chat_id",
			":sender_id",
		}).BindMap(map[string]interface{}{
		":message":    message.Message,
		":updated_at": message.UpdatedAt,
		":id":         message.ID,
		":chat_id":    message.ChatID,
		":sender_id":  message.SenderID,
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

func (r *RepositoryMessenger) DeleteMessage(ctx context.Context, messageID snowflake.ID, chatID, senderID uuid.UUID) error {
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

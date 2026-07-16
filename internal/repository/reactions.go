package repository

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) InsertReaction(ctx context.Context, reaction dto.Reaction) error {
	insertReaction := r.Session.ContextQuery(
		ctx,
		`INSERT INTO messenger_keyspace.reactions (chat_id, user_id, message_id, reaction, sended_at) VALUES (?, ?, ?, ?, ?)`,
		[]string{":chat_id", ":user_id", ":message_id", ":reaction", ":sended_at"}).
		BindMap(map[string]interface{}{
			":chat_id":    reaction.ChatID,
			":user_id":    reaction.UserID,
			":message_id": reaction.MessageID,
			":reaction":   reaction.Reaction,
			":sended_at":  reaction.SendedAt,
		})

	if err := insertReaction.ExecRelease(); err != nil {
		r.Log.Error("error with inserting reaction: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectReactions(ctx context.Context, chatID uuid.UUID) ([]dto.Reaction, error) {
	selectReactions := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.reactions WHERE chat_id=?`,
		[]string{":chat_id"}).
		BindMap(map[string]interface{}{":chat_id": chatID})

	var reactions []dto.Reaction
	if err := selectReactions.SelectRelease(&reactions); err != nil {
		r.Log.Error("error with selecting all reactions from chat: ", "error", err)
		return nil, err
	}

	return reactions, nil
}

func (r *RepositoryMessenger) SelectReactionById(ctx context.Context, chatID, messageID, userID uuid.UUID) (dto.Reaction, error) {
	selectReactionById := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.reactions WHERE chat_id=? AND user_id=? AND message_id=?`,
		[]string{":chat_id", ":user_id", ":message_id"}).
		BindMap(map[string]interface{}{
			":chat_id":    chatID,
			":message_id": messageID,
			":user_id":    userID,
		})

	var reaction dto.Reaction
	if err := selectReactionById.SelectRelease(&reaction); err != nil {
		r.Log.Error("error with selecting reaction by id: ", "error", err)
		return dto.Reaction{}, err
	}

	return reaction, nil
}

func (r *RepositoryMessenger) DeleteReaction(ctx context.Context, messageID snowflake.ID, chatID, userID uuid.UUID) error {
	deleteReaction := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.reactions WHERE chat_id=? AND message_id=? AND user_id=?`,
		[]string{":chat_id", ":message_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":chat_id":    chatID,
			":message_id": messageID,
			":user_id":    userID,
		})

	if err := deleteReaction.ExecRelease(); err != nil {
		r.Log.Error("error with deleting reaction: ", "error", err)
		return err
	}

	return nil
}

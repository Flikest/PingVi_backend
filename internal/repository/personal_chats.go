package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) CreatePersonalChat(ctx context.Context, personalChat dto.PersonalChat) (dto.PersonalChat, error) {
	now := time.Now()

	createPersonalChat := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.personal_chats (id, user1_id, user2_id, created_at)
		VALUES (?, ?, ?, ?)
		`,
		[]string{
			":id",
			":user1_id",
			":user2_id",
			":created_at",
		}).
		BindMap(map[string]interface{}{
			":id":         personalChat.ID,
			":user1_id":   personalChat.User1ID,
			":user2_id":   personalChat.User2ID,
			":created_at": now,
		})

	if err := createPersonalChat.ExecRelease(); err != nil {
		r.Log.Error("error with inserting personalChat")
		return dto.PersonalChat{}, err
	}
	personalChat.CreatedAt = now

	return personalChat, nil
}

func (r *RepositoryMessenger) DeletePersonalChat(ctx context.Context, personalChat dto.DeletePersonalChat) (uuid.UUID, error) {
	deletePersonalChat := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.personal_chats WHERE id=?`,
		[]string{
			":id",
		}).
		BindMap(map[string]interface{}{
			":id": personalChat.ID,
		})

	if err := deletePersonalChat.ExecRelease(); err != nil {
		r.Log.Error("error with deleting personal chat: ", "error", err)
		return uuid.Nil, err
	}

	return personalChat.ID, nil
}

func (r *RepositoryMessenger) IsBelongsToChat(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	selectUsersFromPersonalChat := r.Session.ContextQuery(
		ctx,
		`SELECT user1_id, user2_id FROM messenger_keyspace.personal_chats WHERE id=?`,
		[]string{
			":id",
		}).
		BindMap(map[string]interface{}{
			":id": chatID,
		})

	selectResponse := struct {
		user1ID uuid.UUID
		user2ID uuid.UUID
	}{}

	if err := selectUsersFromPersonalChat.SelectRelease(selectResponse); err != nil {
		r.Log.Error("error with selecting user1 and user2 from perosnal chat", "error", err)
		return false, err
	}

	if userID != selectResponse.user1ID || userID != selectResponse.user2ID {
		r.Log.Warn("not enough rights")
		return false, errors.New("not enough rights")
	}

	return true, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
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

func (r *RepositoryMessenger) SelectPersonalChatMemberIdsByUserId(ctx context.Context, userID gocql.UUID) ([]gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user1_id, user2_id WHERE user1_id=? OR user2_id=?`,
		[]string{
			":user1_id",
			":user2_id",
		}).
		BindMap(map[string]interface{}{
			":user1_id": userID,
			":user2_id": userID,
		})

	userIDs := []struct {
		user1ID gocql.UUID
		user2ID gocql.UUID
	}{}
	if err := query.SelectRelease(&userIDs); err != nil {
		r.Log.Error("error with selectiing user1id and user2id from personal chat: ", "error", err)
		return nil, err
	}

	var result []gocql.UUID

	for _, ids := range userIDs {
		if ids.user1ID == userID {
			result = append(result, ids.user2ID)
		} else {
			result = append(result, ids.user1ID)
		}
	}

	return result, nil
}

func (r *RepositoryMessenger) SelectPersonalChats(ctx context.Context, userID gocql.UUID) ([]dto.PersonalChatResponse, error) {
	var chats1, chats2 []dto.PersonalChat

	// Пользователь как user1
	q1 := r.Session.ContextQuery(ctx,
		`SELECT * FROM messenger_keyspace.personal_chats WHERE user1_id = :uid`,
		[]string{":uid"},
	).BindMap(map[string]interface{}{":uid": userID})
	if err := q1.SelectRelease(&chats1); err != nil {
		r.Log.Error("error with selecting personal chats (user1): ", "error", err)
		return nil, err
	}

	// Пользователь как user2
	q2 := r.Session.ContextQuery(ctx,
		`SELECT * FROM messenger_keyspace.personal_chats WHERE user2_id = :uid`,
		[]string{":uid"},
	).BindMap(map[string]interface{}{":uid": userID})
	if err := q2.SelectRelease(&chats2); err != nil {
		r.Log.Error("error with selecting personal chats (user2): ", "error", err)
		return nil, err
	}

	personalChats := append(chats1, chats2...)

	response := make([]dto.PersonalChatResponse, 0, len(personalChats))
	for _, chat := range personalChats {
		var id gocql.UUID
		if chat.User1ID == userID {
			id = chat.User2ID
		} else {
			id = chat.User1ID
		}
		response = append(response, dto.PersonalChatResponse{
			ID:        chat.ID,
			UserID:    id,
			CreatedAt: chat.CreatedAt,
		})
	}

	return response, nil
}

func (r *RepositoryMessenger) DeletePersonalChat(ctx context.Context, personalChat dto.DeletePersonalChat) (gocql.UUID, error) {
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
		return gocql.UUID{}, err
	}

	return personalChat.ID, nil
}

func (r *RepositoryMessenger) IsBelongsToChat(ctx context.Context, userID, chatID gocql.UUID) (bool, error) {
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
		user1ID gocql.UUID
		user2ID gocql.UUID
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

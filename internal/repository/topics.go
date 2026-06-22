package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

// FIXME: переписать все на слоеную архитектуру, сейчас слой овечает и за бизнес логику и за рабуту с бд

func (r *RepositoryMessenger) CreateTopic(ctx context.Context, topic dto.CreateTopic) (dto.Topic, error) {
	permission, err := r.SelectPermissions(ctx, topic.UserID, topic.ChannelId)
	if err != nil {
		r.Log.Error("error with selecting user permission from channel: ", "error", err)
		return dto.Topic{}, err
	}

	if permission[5] != '1' {
		r.Log.Warn("the user does not have enough rights")
		return dto.Topic{}, errors.New("the user does not have enough rights")
	}

	topicID, err := uuid.NewV7()
	if err != nil {
		r.Log.Error("erorr with generating topic id:", "error", err)
		return dto.Topic{}, err
	}

	now := time.Now()

	createTopic := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.topics(id, channel_id, name, topic_type, position, is_private, created_at)
		VALUES (?,?,?,?,?,?,?)
		`,
		[]string{
			":id",
			":channel_id",
			":name",
			":topic_type",
			":position",
			":is_private",
			":created_at",
		}).
		BindMap(map[string]interface{}{
			":id":         topicID,
			":channel_id": topic.ChannelId,
			":name":       topic.Name,
			":topic_type": topic.TopicType,
			":position":   topic.Position,
			":is_private": topic.IsPrivate,
			":created_at": now,
		})

	if err := createTopic.ExecRelease(); err != nil {
		r.Log.Error("error with inserting topic: ", "error", err)
		return dto.Topic{}, err
	}

	return dto.Topic{
		ID:        topicID,
		ChannelId: topic.ChannelId,
		Name:      topic.Name,
		TopicType: topic.TopicType,
		Position:  topic.Position,
		IsPrivate: topic.IsPrivate,
		CreatedAt: now,
	}, nil
}

func (r *RepositoryMessenger) SelectTopicsFromChannel(ctx context.Context, channelID uuid.UUID) ([]dto.Topic, error) {
	selectTopic := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.topics WHERE channel_id=?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	var topics []dto.Topic
	if err := selectTopic.SelectRelease(&topics); err != nil {
		r.Log.Error("error with selecting topics from channel: ", "error", err)
		return nil, err
	}

	return topics, nil
}

func (r *RepositoryMessenger) UpdateTopic(ctx context.Context, topic dto.UpdateTopic) (dto.Topic, error) {
	now := time.Now()

	permision, err := r.SelectPermissions(ctx, topic.UserID, topic.ChannelId)
	if err != nil {
		r.Log.Error("error with selecting user permision from channel: ", "error", err)
		return dto.Topic{}, err
	}

	if permision[5] != '1' {
		return dto.Topic{}, errors.New("the user does not have enough rights")
	}

	updateTopic := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.topics
		SET name=?, topic_type=?, position=?, is_private=?, updated_at=? 
		WHERE id=? AND channel_id=?
		`,
		[]string{
			":name",
			":topic_type",
			":position",
			":is_private",
			":updated_at",
			":id",
			":channel_id",
		}).BindMap(map[string]interface{}{
		":name":       topic.Name,
		":topic_type": topic.TopicType,
		":position":   topic.Position,
		":is_private": topic.IsPrivate,
		":updated_at": now,
		":id":         topic.ID,
		":channel_id": topic.ChannelId,
	})

	if err := updateTopic.ExecRelease(); err != nil {
		r.Log.Error("error with updating topic: ", "error", err)
		return dto.Topic{}, err
	}

	return dto.Topic{
		ID:        topic.ID,
		ChannelId: topic.ChannelId,
		Name:      topic.Name,
		TopicType: topic.TopicType,
		Position:  topic.Position,
		IsPrivate: topic.IsPrivate,
		UpdatedAt: now,
	}, nil
}

func (r *RepositoryMessenger) DeleteTopic(ctx context.Context, topic dto.DeleteTopic) (uuid.UUID, error) {
	permission, err := r.SelectPermissions(ctx, topic.UserId, topic.ChannelId)
	if err != nil {
		r.Log.Error("error with selecting user permission: ", "error", err)
		return uuid.Nil, err
	}

	if permission[5] != '1' {
		r.Log.Warn("the user does not have enough rights")
		return uuid.Nil, errors.New("the user does not have enough rights")
	}

	deleteMessages := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.messages WHERE chat_id=?`,
		[]string{
			":chat_id",
		}).
		BindMap(map[string]interface{}{
			":chat_id": topic.ID,
		})

	if err := deleteMessages.ExecRelease(); err != nil {
		r.Log.Error("error with deleting message: ", "error", err)
		return uuid.Nil, err
	}

	deleteTopic := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.topics WHERE id=? AND channel_id=?`,
		[]string{
			":id",
			":channel_id",
		}).
		BindMap(map[string]interface{}{
			":id":         topic.ID,
			":channel_id": topic.ChannelId,
		})
	if err := deleteTopic.ExecRelease(); err != nil {
		r.Log.Error("error with deleting topic: ", "error", err)
		return uuid.Nil, err
	}

	return topic.ID, nil
}

package repository

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) SelectChannelOwnerID(ctx context.Context, channelID uuid.UUID) (uuid.UUID, error) {
	SelectOwnerID := r.Session.ContextQuery(
		ctx,
		`SELECT owner_id FROM messenger_keyspace.channels WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": channelID,
		})

	var ownerID uuid.UUID
	if err := SelectOwnerID.SelectRelease(&ownerID); err != nil {
		r.Log.Error("error with selecting owner id from channel: ", "error", err)
		return uuid.Nil, err
	}

	return ownerID, nil
}

func (r *RepositoryMessenger) InsertChannel(ctx context.Context, channel dto.Channel) error {
	createChannel := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.channels(id, name, description, owner_id, icon_url, is_public, created_at)
		VALUE (?,?,?,?,?,?,?)
		`,
		[]string{
			":id",
			":name",
			":description",
			":owner_id",
			":icon_url",
			":is_public",
			":created_at",
		}).BindMap(map[string]interface{}{
		":id":          channel.ID,
		":name":        channel.Name,
		":description": channel.Description,
		":owner_id":    channel.OwnerID,
		":icon_url":    channel.IconURL,
		":is_public":   channel.IsPublic,
		":created_at":  channel.CreatedAt,
	})

	if err := createChannel.ExecRelease(); err != nil {
		r.Log.Error("error with inserting channel: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) InsertChannelMember(ctx context.Context, member dto.ChannelMember) error {
	insertMember := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.channel_members (channel_id, user_id, role_id, joined_at)
		VALUES (?, ?, ?, ?)
		`,
		[]string{
			":channel_id",
			":user_id",
			":role_id",
			":joined_at",
		}).
		BindMap(map[string]interface{}{
			":channel_id": member.ChannelID,
			":user_id":    member.UserID,
			":role_id":    member.RoleID,
			":joined_at":  member.JoinedAt,
		})

	if err := insertMember.ExecRelease(); err != nil {
		r.Log.Error("error with inserting member: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectChannels(ctx context.Context, userID uuid.UUID) ([]dto.Channel, error) {
	selectMembers := r.Session.ContextQuery(
		ctx,
		`SELECT channel_id FROM messenger_keyspace.channel_members WHERE user_id IN ?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var userIDs []string
	if err := selectMembers.SelectRelease(&userIDs); err != nil {
		r.Log.Error("error with selecting members by id: ", "error", err)
		return nil, err
	}

	query := r.Session.Query(
		`SELECT * FROM messenger_keyspace.channels WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": userIDs,
		})

	var channels []dto.Channel
	if err := query.SelectRelease(&channels); err != nil {
		r.Log.Error("error with selecting channels: ", "error", err)
		return nil, err
	}

	return channels, nil
}

func (r *RepositoryMessenger) UpdateChannel(ctx context.Context, channel dto.Channel) error {
	updateChannel := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.channels
		SET name=?, description=?, icon_url=?, is_public=?, updated_at=?
		WHERE id=?
		`, []string{
			":name",
			":description",
			":icon_url",
			":is_public",
			":updated_at",
			":id"}).
		BindMap(map[string]interface{}{
			":name":        channel.Name,
			":description": channel.Description,
			":icon_url":    channel.IconURL,
			":is_public":   channel.IsPublic,
			":updated_at":  channel.UpdatedAt,
			":id":          channel.ID,
		})

	if err := updateChannel.ExecRelease(); err != nil {
		r.Log.Error("error with updating channel: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelMemberByUserID(ctx context.Context, channelID, userID uuid.UUID) error {
	deleteMember := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.channel_members WHERE user_id=? AND channel_id=?`,
		[]string{
			":user_id",
			":channel_id",
		}).
		BindMap(map[string]interface{}{
			":user_id":    userID,
			":channel_id": channelID,
		})

	if err := deleteMember.ExecRelease(); err != nil {
		r.Log.Error("error with deleting member from channel: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelMember(ctx context.Context, channelID, memberID uuid.UUID) error {
	deleteMember := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.channel_members WHERE channel_id=? and user_id=?`,
		[]string{
			":channel_id",
			":user_id",
		}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
			":user_id":    memberID,
		})

	if err := deleteMember.ExecRelease(); err != nil {
		r.Log.Error("error with deleting member from channel: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelRolesByChannelID(ctx context.Context, channelID uuid.UUID) error {
	deleteRole := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.channel_roles WHERE channel_id = ?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	if err := deleteRole.ExecRelease(); err != nil {
		r.Log.Error("error with deleting roles from chat: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelTopicsByChannelID(ctx context.Context, channelID uuid.UUID) error {
	deleteTopics := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.topics WHERE channel_id=?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	if err := deleteTopics.ExecRelease(); err != nil {
		r.Log.Error("error with deleting topics from channel: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelMembersByChannelID(ctx context.Context, channelID uuid.UUID) error {
	deleteMember := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.channel_members WHERE channel_id=?`,
		[]string{
			":channel_id",
		}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	if err := deleteMember.ExecRelease(); err != nil {
		r.Log.Error("error with deleting members from chat: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteChannelByID(ctx context.Context, channelID uuid.UUID) error {
	deleteChannel := r.Session.ContextQuery(
		ctx,
		`
		DELETE FROM messenger_keyspace.channels WHERE id=?
		`,
		[]string{
			":id",
		}).
		BindMap(map[string]interface{}{
			":id": channelID,
		})

	if err := deleteChannel.ExecRelease(); err != nil {
		r.Log.Error("error with deleting channel: ", "error", err)
		return err
	}

	return nil
}

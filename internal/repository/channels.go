package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
)

func toInterfaces(ids []gocql.UUID) []interface{} {
	out := make([]interface{}, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func (r *RepositoryMessenger) IsChannelMuted(ctx context.Context, channelID, userID gocql.UUID) (bool, error) {
	selectMuted := r.Session.ContextQuery(
		ctx,
		`SELECT is_muted FROM messenger_keyspace.channel_members WHERE channel_id=? AND user_id=?`,
		[]string{":channel_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
			":user_id":    userID,
		})

	var isMuted bool
	if err := selectMuted.Scan(&isMuted); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return false, nil
		}
		r.Log.Error("error with selecting is_muted from channel members", "error", err,
			"channel_id", channelID, "user_id", userID)
		return false, err
	}

	return isMuted, nil
}

func (r *RepositoryMessenger) SelectChannelOwnerID(ctx context.Context, channelID gocql.UUID) (gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT owner_id FROM messenger_keyspace.channels WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": channelID,
		})

	var ownerID gocql.UUID
	if err := query.Scan(&ownerID); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return gocql.UUID{}, err
		}
		r.Log.Error("error with selecting owner id from channel", "error", err,
			"channel_id", channelID)
		return gocql.UUID{}, err
	}

	return ownerID, nil
}

func (r *RepositoryMessenger) SelectChannelIdsByUserId(ctx context.Context, userID gocql.UUID) ([]gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT channel_id FROM messenger_keyspace.channel_members WHERE user_id=?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var chatsID []gocql.UUID
	if err := query.SelectRelease(&chatsID); err != nil {
		r.Log.Error("error with selecting chats by user id: ", "error", err)
		return nil, err
	}

	return chatsID, nil
}

func (r *RepositoryMessenger) SelectChannelMemberIds(ctx context.Context, id gocql.UUID) ([]gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user_id FROM messenger_keyspace.channel_members WHERE channel_id=?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": id,
		})

	var membersID []gocql.UUID
	if err := query.SelectRelease(&membersID); err != nil {
		r.Log.Error("error with selecting members by channel id: ", "error", err)
		return nil, err
	}

	return membersID, nil
}

func (r *RepositoryMessenger) SelectChannelMembers(ctx context.Context, channelID gocql.UUID) ([]dto.ChannelMember, error) {
	query := r.Session.ContextQuery(ctx,
		`SELECT channel_id, user_id, role_id, joined_at, is_muted
		 FROM messenger_keyspace.channel_members WHERE channel_id = ?`,
		[]string{":channel_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
		})

	var members []dto.ChannelMember
	if err := query.SelectRelease(&members); err != nil {
		r.Log.Error("error with selecting channel members", "error", err,
			"channel_id", channelID)
		return nil, err
	}
	return members, nil
}

func (r *RepositoryMessenger) InsertChannel(ctx context.Context, channel dto.Channel) error {
	createChannel := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.channels(id, name, description, owner_id, icon_urls, is_public, created_at)
		VALUES (?,?,?,?,?,?,?)
		`,
		[]string{
			":id",
			":name",
			":description",
			":owner_id",
			":icon_urls",
			":is_public",
			":created_at",
		}).BindMap(map[string]interface{}{
		":id":          channel.ID,
		":name":        channel.Name,
		":description": channel.Description,
		":owner_id":    channel.OwnerID,
		":icon_urls":   channel.IconURLs,
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

func (r *RepositoryMessenger) IsBelongChannel(ctx context.Context, channelID, userID gocql.UUID) (bool, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user_id FROM messenger_keyspace.channel_members WHERE channel_id = ? AND user_id = ? LIMIT 1`,
		[]string{":channel_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
			":user_id":    userID,
		})

	var scanUserID gocql.UUID
	if err := query.Scan(&scanUserID); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return false, nil
		}
		r.Log.Error("error with checking belonging to channel", "error", err,
			"channel_id", channelID, "user_id", userID)
		return false, err
	}

	return true, nil
}

func (r *RepositoryMessenger) SelectChannelNameByID(ctx context.Context, channelID gocql.UUID) (string, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT name FROM messenger_keyspace.channels WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": channelID,
		})

	var name string
	if err := query.Scan(&name); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return "", err
		}
		r.Log.Error("error with selecting channel name", "error", err,
			"channel_id", channelID)
		return "", err
	}

	return name, nil
}

func (r *RepositoryMessenger) SelectChannels(ctx context.Context, userID gocql.UUID) ([]dto.Channel, error) {
	selectMembers := r.Session.ContextQuery(
		ctx,
		`SELECT channel_id FROM messenger_keyspace.channel_members WHERE user_id = ?`,
		nil,
	).Bind(userID)

	var channelIDs []gocql.UUID
	if err := selectMembers.SelectRelease(&channelIDs); err != nil {
		r.Log.Error("error with selecting channel members by user id",
			"error", err,
			"user_id", userID,
		)
		return nil, err
	}

	if len(channelIDs) == 0 {
		return []dto.Channel{}, nil
	}

	placeholders := make([]string, len(channelIDs))
	args := make([]interface{}, len(channelIDs))
	for i, id := range channelIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	queryStr := fmt.Sprintf(
		`SELECT id, name, description, owner_id, icon_urls, is_public, created_at, updated_at
		 FROM messenger_keyspace.channels
		 WHERE id IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	query := r.Session.ContextQuery(ctx, queryStr, nil).Bind(args...)

	var channels []dto.Channel
	if err := query.SelectRelease(&channels); err != nil {
		r.Log.Error("error with selecting channels by ids",
			"error", err,
			"user_id", userID,
			"channel_ids_count", len(channelIDs),
		)
		return nil, err
	}

	return channels, nil
}

func (r *RepositoryMessenger) UpdateChannel(ctx context.Context, channel dto.Channel) error {
	updateChannel := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.channels
		SET name=?, description=?, icon_urls=?, is_public=?, updated_at=?
		WHERE id=?
		`, []string{
			":name",
			":description",
			":icon_urls",
			":is_public",
			":updated_at",
			":id"}).
		BindMap(map[string]interface{}{
			":name":        channel.Name,
			":description": channel.Description,
			":icon_urls":   channel.IconURLs,
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

func (r *RepositoryMessenger) DeleteChannelMemberByUserID(ctx context.Context, channelID, userID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteChannelMember(ctx context.Context, channelID, memberID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteChannelRolesByChannelID(ctx context.Context, channelID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteChannelTopicsByChannelID(ctx context.Context, channelID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteChannelMembersByChannelID(ctx context.Context, channelID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteChannelByID(ctx context.Context, channelID gocql.UUID) error {
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

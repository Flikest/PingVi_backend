package repository

import (
	"context"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) SelectPermissions(ctx context.Context, userID, channelID uuid.UUID) (string, error) {
	selectUserRole := r.Session.ContextQuery(
		ctx,
		`SELECT role_id FROM messenger_keyspace.channel_members WHERE user_id=? AND channel_id=?`,
		[]string{":user_id", ":channel_id"}).BindMap(map[string]interface{}{
		":user_id":    userID,
		":channel_id": channelID,
	})

	var userRoleID string
	if err := selectUserRole.SelectRelease(&userRoleID); err != nil {
		r.Log.Error("error with selecting user role id: ", "error", err)
		return "", err
	}

	selectPermissions := r.Session.ContextQuery(
		ctx,
		`SELECT permissions FROM messenger_keyspace.channel_roles WHERE id=? AND channel_id=?`,
		[]string{":id", ":user_id"}).BindMap(map[string]interface{}{
		":id":         userRoleID,
		":channel_id": channelID,
	})

	var Permissions string
	if err := selectPermissions.SelectRelease(&Permissions); err != nil {
		r.Log.Error("error with selecting permissions: ", "error", err)
		return "", err
	}

	return Permissions, nil
}

func (r *RepositoryMessenger) InsertChannelRole(ctx context.Context, role dto.Role) error {
	createRole := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.channel_roles (id, channel_id, name, color, permissions, is_mentionable, is_default, created_at)
		VALUES (?,?,?,?,?,?,?,?)
		`, []string{
			":id",
			":channel_id",
			":name",
			":color",
			":permissions",
			":is_mentionable",
			":is_default",
			":created_at",
		}).BindMap(map[string]interface{}{
		":id":             role.ID,
		":channel_id":     role.ChannelID,
		":name":           role.Name,
		":color":          role.Color,
		":permissions":    role.Permissions,
		":is_mentionable": role.IsMentionable,
		":is_default":     role.IsDefault,
		":created_at":     role.CreatedAt,
	})

	if err := createRole.ExecRelease(); err != nil {
		r.Log.Error("error with inserting role", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectRoles(ctx context.Context, channelID uuid.UUID) ([]dto.Role, error) {
	selectRoles := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.channel_roles WHERE channel_id=?`,
		[]string{":channel_id:"}).
		BindMap(map[string]interface{}{":channel_id": channelID})

	var roles []dto.Role
	if err := selectRoles.SelectRelease(&roles); err != nil {
		r.Log.Error("error with selecting roles from channel: ", "error", err)
		return nil, err
	}

	return roles, nil
}

func (r *RepositoryMessenger) SelectRoleName(ctx context.Context, roleID, channelID uuid.UUID) (string, error) {
	selectRole := r.Session.ContextQuery(
		ctx,
		`SELECT name FROM messenger_keyspace.channel_roles WHERE id=? AND channel_id=?`,
		[]string{":id", ":channel_id"}).
		BindMap(map[string]interface{}{
			":id":         channelID,
			":channel_id": channelID,
		})

	var name string
	if err := selectRole.SelectRelease(&name); err != nil {
		r.Log.Error("error with selecting role name: ", "error", err)
		return "", err
	}

	return name, nil
}

func (r *RepositoryMessenger) SelectMemberIDsByRoleID(ctx context.Context, channelID, roleID uuid.UUID) ([]uuid.UUID, error) {
	selectUserIDs := r.Session.ContextQuery(
		ctx,
		`
		SELECT user_id FROM messenger_keyspace.channel_members WHERE channel_id=? AND role_id=?`,
		[]string{":channel_id", ":role_id"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
			":role_id":    roleID,
		})

	var userIDs []uuid.UUID

	if err := selectUserIDs.SelectRelease(&userIDs); err != nil {
		r.Log.Error("error with selecting user ids by role id: ", "error", err)
		return nil, err
	}

	return userIDs, nil
}

func (r *RepositoryMessenger) UpdateRole(ctx context.Context, role dto.UpdateRole) (dto.Role, error) {

	now := time.Now()

	updateRole := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.channel_roles
		SET name=?, color=?, permissions=?, is_mentionable=?, is_default=?, updated_at=?
		WHERE id=? AND channel_id=?
		`, []string{
			":name",
			":color",
			":permissions",
			":is_mentionable",
			":is_default",
			":updated_at",
			":id",
			":channel_id",
		}).
		BindMap(map[string]interface{}{
			":name":           role.Name,
			":color":          role.Color,
			":permissions":    role.Permissions,
			":is_mentionable": role.IsMentionable,
			":is_default":     role.IsDefault,
			":updated_at":     now,
			":id":             role.ID,
			":channel_id":     role.ChannelID,
		})

	if err := updateRole.ExecRelease(); err != nil {
		r.Log.Error("error with updating role from channel: ", "error", err)
		return dto.Role{}, err
	}

	return dto.Role{
		ID:            role.ID,
		ChannelID:     role.ChannelID,
		Name:          role.Name,
		Color:         role.Color,
		Permissions:   role.Permissions,
		IsMentionable: role.IsMentionable,
		IsDefault:     role.IsDefault,
		UpdatedAt:     now,
	}, nil
}

func (r *RepositoryMessenger) SelectEveryoneRoleID(ctx context.Context, channelID uuid.UUID) (uuid.UUID, error) {
	selectEveryoneRoleID := r.Session.ContextQuery(
		ctx,
		`SELECT id FROM messenger_keyspace.channel_roles WHERE channel_id=? AND name=?`,
		[]string{":channel_id", ":name"}).
		BindMap(map[string]interface{}{
			":channel_id": channelID,
			":name":       "@everyone",
		})

	var everyoneRoleID uuid.UUID
	if err := selectEveryoneRoleID.SelectRelease(&everyoneRoleID); err != nil {
		r.Log.Error("error with selecting role id on everyone name: ", "error", err)
		return uuid.Nil, err
	}

	return everyoneRoleID, nil
}

func (r *RepositoryMessenger) UpdateChannelMemberRole(ctx context.Context, channelID, userID, roleID uuid.UUID) error {
	updateUserRole := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.channel_members
		SET role_id = ?
		WHERE channel_id=? AND user_id IN ?
		`, []string{":role_id", ":channel_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":role_id":    roleID,
			":channel_id": channelID,
			":user_id":    userID,
		})

	if err := updateUserRole.ExecRelease(); err != nil {
		r.Log.Error("error with updating user roles: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) UpdateChannelMembersRole(ctx context.Context, channelID, roleID uuid.UUID, userIDs []uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}

	batch := r.Session.NewBatch(gocql.LoggedBatch)
	query := `UPDATE messenger_keyspace.channel_members SET role_id = ? WHERE channel_id = ? AND user_id = ?`

	for _, userID := range userIDs {
		batch.Query(query, roleID, channelID, userID)
	}

	return r.Session.ExecuteBatch(batch.WithContext(ctx))
}

func (r *RepositoryMessenger) DeleteChannelRoleByID(ctx context.Context, roleID, channelID uuid.UUID) error {
	deleteRole := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.channel_roles WHERE id=? AND channel_id=?`,
		[]string{":id", ":channel_id"}).
		BindMap(map[string]interface{}{
			":id":         roleID,
			":channel_id": channelID,
		})

	if err := deleteRole.ExecRelease(); err != nil {
		r.Log.Error("error with deleting role: ", "error", err)
		return err
	}

	return nil
}

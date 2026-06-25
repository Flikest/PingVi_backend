package repository

import (
	"context"
	"errors"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) SelectGroupOwnerID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error) {
	selectOwnerID := r.Session.ContextQuery(
		ctx,
		`SELECT owner_id FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": groupID,
		})

	var ownerID uuid.UUID
	if err := selectOwnerID.SelectRelease(&ownerID); err != nil {
		r.Log.Error("error with selecting owner id: ", "error", err)
		return uuid.Nil, err
	}

	return ownerID, nil
}

func (r *RepositoryMessenger) SelectGroupMemberIsAdmin(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	selectMemberIsAdmin := r.Session.ContextQuery(
		ctx,
		`SELECT is_admin FROM messenger_keyspace.group_members WHERE group_id=? AND user_id=?`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var isAdmin bool
	if err := selectMemberIsAdmin.SelectRelease(&isAdmin); err != nil {
		r.Log.Error("error with selecting is_admin: ", "error", err)
		return false, err
	}

	return isAdmin, nil
}

func (r *RepositoryMessenger) SelectGroupExists(ctx context.Context, groupID uuid.UUID) (bool, error) {
	q := r.Session.ContextQuery(
		ctx,
		`SELECT COUNT(*) FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).BindMap(map[string]interface{}{
		":id": groupID,
	})

	var count int
	if err := q.SelectRelease(&count); err != nil {
		r.Log.Error("error with selecting count group by id: ", "error", err)
		return false, err
	}

	return count > 0, nil
}

func (r *RepositoryMessenger) InsertGroup(ctx context.Context, group dto.Group) error {
	createGroup := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.groups (id, name, avatar_links, owner_id, is_public, created_at)
		VALUES (?,?,?,?,?,?)
		`, []string{
			":id",
			":name",
			":avatar_links",
			":owner_id",
			":is_public",
			":created_at",
		}).BindMap(map[string]interface{}{
		":id":           group.ID,
		":name":         group.Name,
		":avatar_links": group.AvatarLinks,
		":owner_id":     group.OwnerID,
		":is_public":    group.IsPublic,
		":created_at":   group.CreatedAt,
	})

	if err := createGroup.ExecRelease(); err != nil {
		r.Log.Error("error with creating group: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) InsertGroupMember(ctx context.Context, member dto.GroupMember) error {
	insertMember := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.group_members (group_id, user_id, is_admin, joined_at)
		VALUES (?,?,?,?)
		`,
		[]string{
			":group_id",
			":user_id",
			":is_admin",
			":joined_at",
		}).
		BindMap(map[string]interface{}{
			":group_id":  member.GroupID,
			":user_id":   member.UserID,
			":is_admin":  member.IsAdmin,
			":joined_at": member.JoinedAt,
		})

	if err := insertMember.ExecRelease(); err != nil {
		r.Log.Error("error with inserting members to group: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	deleteMember := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.group_members WHERE group_id=? AND user_id=?`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	if err := deleteMember.ExecRelease(); err != nil {
		r.Log.Error("error with deleting member: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) SelectAllMembersGroup(ctx context.Context, groupID uuid.UUID) ([]dto.GroupMember, error) {
	selectMembersFromGroup := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.group_members WHERE group_id=?`,
		[]string{
			":group_id",
		}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
		})

	var members []dto.GroupMember
	if err := selectMembersFromGroup.SelectRelease(&members); err != nil {
		r.Log.Error("error with selecting members from group: ", "error", err)
		return nil, err
	}

	return members, nil
}

func (r *RepositoryMessenger) SelectGroupNameByID(ctx context.Context, groupID uuid.UUID) (string, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT name FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": groupID,
		})

	var name string
	if err := query.SelectRelease(name); err != nil {
		r.Log.Error("error with selecting group name by id: ", "error", err)
		return "", err
	}

	return name, nil
}

func (r *RepositoryMessenger) IsBelongGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user_id FROM messenger_keyspace.group_members WHERE group_id = ? AND user_id = ? LIMIT 1`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var scanUserID uuid.UUID
	if err := query.SelectRelease(&scanUserID); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return false, nil
		}

		r.Log.Error("error with checking belonging to group: ", "error", err)
		return false, err
	}

	return true, nil
}

func (r *RepositoryMessenger) SelectMemberGroup(ctx context.Context, groupID, userID uuid.UUID) (dto.GroupMember, error) {
	selectMemberFromGroup := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.group_members WHERE group_id=? AND user_id=?`,
		[]string{
			":group_id",
			":user_id",
		}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var member dto.GroupMember
	if err := selectMemberFromGroup.SelectRelease(&member); err != nil {
		r.Log.Error("error with selecting member from group: ", "error", err)
		return dto.GroupMember{}, err
	}

	return member, nil
}

func (r *RepositoryMessenger) SelectGroup(ctx context.Context, userID uuid.UUID) ([]dto.Group, error) {
	selectMembers := r.Session.ContextQuery(
		ctx,
		`SELECT group_id FROM messenger_keyspace.groups WHERE user_id=?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var groupIDs []uuid.UUID
	if err := selectMembers.SelectRelease(&groupIDs); err != nil {
		r.Log.Error("error with selecting groups id: ", "error", err)
		return nil, err
	}

	if len(groupIDs) == 0 {
		return []dto.Group{}, nil
	}

	var groups []dto.Group
	query := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.groups WHERE group_id IN ?`,
		[]string{":group_ids"}).
		BindMap(map[string]interface{}{
			":group_ids": groupIDs,
		})

	if err := query.SelectRelease(&groups); err != nil {
		r.Log.Error("error with selecting groups by ids: ", "error", err)
		return nil, err
	}

	return groups, nil
}

func (r *RepositoryMessenger) UpdateGroup(ctx context.Context, group dto.Group) error {
	updateGroup := r.Session.ContextQuery(
		ctx,
		`
		UPDATE messenger_keyspace.groups
		SET name=?, avatar_links=?, is_public=?, updated_at=?
		WHERE id=? AND owner_id=?
		`, []string{
			":name",
			":avatar_links",
			":is_public",
			":updated_at",
			":id",
			":owner_id",
		}).BindMap(map[string]interface{}{
		":name":         group.Name,
		":avatar_links": group.AvatarLinks,
		":is_public":    group.IsPublic,
		":updated_at":   group.UpdatedAt,
		":id":           group.ID,
		":owner_id":     group.OwnerID,
	})

	if err := updateGroup.ExecRelease(); err != nil {
		r.Log.Error("error with updating group: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteGroupMembersByGroupID(ctx context.Context, groupID uuid.UUID) error {
	deleteMembers := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.group_members WHERE group_id=?`,
		[]string{":group_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
		})

	if err := deleteMembers.ExecRelease(); err != nil {
		r.Log.Error("error with deleting members: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryMessenger) DeleteGroupByID(ctx context.Context, groupID uuid.UUID) error {
	deleteGroup := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.groups WHERE id=?`,
		[]string{
			":id",
		}).BindMap(map[string]interface{}{
		":id": groupID,
	})

	if err := deleteGroup.ExecRelease(); err != nil {
		r.Log.Error("error with deleting group", "error", err)
		return err
	}

	return nil
}

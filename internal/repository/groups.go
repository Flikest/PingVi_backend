package repository

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
)

func (r *RepositoryMessenger) IsGroupMuted(ctx context.Context, groupID, userID gocql.UUID) (bool, error) {
	selectMuted := r.Session.ContextQuery(
		ctx,
		`SELECT is_muted FROM messenger_keyspace.group_members WHERE group_id=? AND user_id=?`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var isMuted []bool
	if err := selectMuted.SelectRelease(&isMuted); err != nil {
		r.Log.Error("error with selecting owner id: ", "error", err)
		return false, err
	}

	if len(isMuted) == 0 {
		return false, nil
	}

	return isMuted[0], nil
}

func (r *RepositoryMessenger) SelectGroupOwnerID(ctx context.Context, groupID gocql.UUID) (gocql.UUID, error) {
	selectOwnerID := r.Session.ContextQuery(
		ctx,
		`SELECT owner_id FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": groupID,
		})

	var ownerIDs []gocql.UUID
	if err := selectOwnerID.SelectRelease(&ownerIDs); err != nil {
		r.Log.Error("error with selecting owner id: ", "error", err)
		return gocql.UUID{}, err
	}

	if len(ownerIDs) == 0 {
		return gocql.UUID{}, gocql.ErrNotFound
	}

	return ownerIDs[0], nil
}

func (r *RepositoryMessenger) SelectGroupIdsByUserId(ctx context.Context, userID gocql.UUID) ([]gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT group_id FROM messenger_keyspace.group_members WHERE user_id=?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var groupIDs []gocql.UUID
	if err := query.SelectRelease(&groupIDs); err != nil {
		r.Log.Error("error with selecting user ids from group: ", "error", err)
		return nil, err
	}

	return groupIDs, nil
}

func (r *RepositoryMessenger) SelectGroupMembers(ctx context.Context, groupID gocql.UUID) ([]gocql.UUID, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user_id FROM messenger_keyspace.group_members WHERE group_id=?`,
		[]string{":group_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
		})

	var userIDs []gocql.UUID
	if err := query.SelectRelease(&userIDs); err != nil {
		r.Log.Error("error with selecting members from group")
		return nil, err
	}

	return userIDs, nil
}

func (r *RepositoryMessenger) SelectGroupMemberIsAdmin(ctx context.Context, groupID, userID gocql.UUID) (bool, error) {
	selectMemberIsAdmin := r.Session.ContextQuery(
		ctx,
		`SELECT is_admin FROM messenger_keyspace.group_members WHERE group_id=? AND user_id=?`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var isAdmin []bool
	if err := selectMemberIsAdmin.SelectRelease(&isAdmin); err != nil {
		r.Log.Error("error with selecting is_admin: ", "error", err)
		return false, err
	}

	if len(isAdmin) == 0 {
		return false, nil
	}

	return isAdmin[0], nil
}

func (r *RepositoryMessenger) SelectGroupExists(ctx context.Context, groupID gocql.UUID) (bool, error) {
	q := r.Session.ContextQuery(
		ctx,
		`SELECT COUNT(*) FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).BindMap(map[string]interface{}{
		":id": groupID,
	})

	var counts []int64
	if err := q.SelectRelease(&counts); err != nil {
		r.Log.Error("error with selecting count group by id: ", "error", err)
		return false, err
	}

	if len(counts) == 0 {
		return false, nil
	}

	return counts[0] > 0, nil
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

func (r *RepositoryMessenger) DeleteGroupMember(ctx context.Context, groupID, userID gocql.UUID) error {
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

func (r *RepositoryMessenger) SelectAllMembersGroup(ctx context.Context, groupID gocql.UUID) ([]dto.GroupMember, error) {
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

func (r *RepositoryMessenger) SelectGroupNameByID(ctx context.Context, groupID gocql.UUID) (string, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT name FROM messenger_keyspace.groups WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": groupID,
		})

	var names []string
	if err := query.SelectRelease(&names); err != nil {
		r.Log.Error("error with selecting group name by id: ", "error", err)
		return "", err
	}

	if len(names) == 0 {
		return "", gocql.ErrNotFound
	}

	return names[0], nil
}

func (r *RepositoryMessenger) IsBelongGroup(ctx context.Context, groupID, userID gocql.UUID) (bool, error) {
	query := r.Session.ContextQuery(
		ctx,
		`SELECT user_id FROM messenger_keyspace.group_members WHERE group_id = ? AND user_id = ? LIMIT 1`,
		[]string{":group_id", ":user_id"}).
		BindMap(map[string]interface{}{
			":group_id": groupID,
			":user_id":  userID,
		})

	var userIDs []gocql.UUID
	if err := query.SelectRelease(&userIDs); err != nil {
		r.Log.Error("error with checking belonging to group: ", "error", err)
		return false, err
	}

	return len(userIDs) > 0, nil
}

func (r *RepositoryMessenger) SelectMemberGroup(ctx context.Context, groupID, userID gocql.UUID) (dto.GroupMember, error) {
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

	var members []dto.GroupMember
	if err := selectMemberFromGroup.SelectRelease(&members); err != nil {
		r.Log.Error("error with selecting member from group: ", "error", err)
		return dto.GroupMember{}, err
	}

	if len(members) == 0 {
		return dto.GroupMember{}, gocql.ErrNotFound
	}

	return members[0], nil
}

func (r *RepositoryMessenger) SelectGroup(ctx context.Context, userID gocql.UUID) ([]dto.Group, error) {
	selectMembers := r.Session.ContextQuery(
		ctx,
		`SELECT group_id FROM messenger_keyspace.group_members WHERE user_id=?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var groupIDs []gocql.UUID
	if err := selectMembers.SelectRelease(&groupIDs); err != nil {
		r.Log.Error("error with selecting groups id: ", "error", err)
		return nil, err
	}

	if len(groupIDs) == 0 {
		return []dto.Group{}, nil
	}

	groups := make([]dto.Group, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		query := r.Session.ContextQuery(
			ctx,
			`SELECT id, name, avatar_links, owner_id, is_public, created_at, updated_at
             FROM messenger_keyspace.groups WHERE id=?`,
			[]string{":id"}).
			BindMap(map[string]interface{}{
				":id": groupID,
			})

		var groupList []dto.Group
		if err := query.SelectRelease(&groupList); err != nil {
			r.Log.Error("error with selecting group by id: ", "error", err)
			return nil, err
		}

		if len(groupList) == 0 {
			r.Log.Warn("group not found, skipping", "group_id", groupID)
			continue
		}

		groups = append(groups, groupList[0])
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

func (r *RepositoryMessenger) DeleteGroupMembersByGroupID(ctx context.Context, groupID gocql.UUID) error {
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

func (r *RepositoryMessenger) DeleteGroupByID(ctx context.Context, groupID gocql.UUID) error {
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

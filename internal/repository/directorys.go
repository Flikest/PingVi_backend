package repository

import (
	"context"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

func (r *RepositoryMessenger) CreateDirectory(ctx context.Context, dir dto.CreateDirectory) (dto.Directory, error) {
	now := time.Now()

	dirID, err := uuid.NewV7()
	if err != nil {
		r.Log.Error("error with generate dir id: ", "error", err)
		return dto.Directory{}, err
	}

	createDirectory := r.Session.ContextQuery(
		ctx,
		`
		INSERT INTO messenger_keyspace.directory (id, channel_id, topic_ids, name, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, []string{":id", ":channel_id", ":topic_ids", ":name", ":created_at"}).
		BindMap(map[string]interface{}{
			":id":         dirID,
			":channel_id": dir.ChannelID,
			":name":       dir.Name,
			":created_at": now,
		})

	if err := createDirectory.ExecRelease(); err != nil {
		r.Log.Error("error when create directory: ", "error", err)
		return dto.Directory{}, err
	}

	return dto.Directory{
		ChannelID: dir.ChannelID,
		Name:      dir.Name,
		CreatedAt: now,
	}, nil
}

func (r *RepositoryMessenger) SelectAllDirectorys(ctx context.Context, channel_id uuid.UUID) ([]dto.Directory, error) {
	selectDirectorys := r.Session.ContextQuery(
		ctx,
		`SELECT * FROM messenger_keyspace.directory WHERE channel_id=?`,
		[]string{":channel_id"}).BindMap(map[string]interface{}{
		":channel_id": channel_id,
	})

	var directorys []dto.Directory
	if err := selectDirectorys.SelectRelease(&directorys); err != nil {
		r.Log.Error("error when select all directorys: ", "error", err)
		return nil, err
	}

	return directorys, nil
}

func (r *RepositoryMessenger) UpdateDirectory(ctx context.Context, dir dto.UpdateDirectory) (dto.Directory, error) {
	now := time.Now()

	updateDirectory := r.Session.ContextQuery(
		ctx,
		`
		UPDATE FROM messenger_keyspace.directory
		SET name=?, updated_at=? 
		WHERE channel_id=? AND id=?
		`,
		[]string{":name", ":updated_at", ":channel_id", ":id"},
	).BindMap(
		map[string]interface{}{
			":name":       dir.Name,
			":updated_at": now,
			":channel_id": dir.ChannelID,
			":id":         dir.ID,
		},
	)

	if err := updateDirectory.ExecRelease(); err != nil {
		r.Log.Error("error when update directory: ", "error", err)
		return dto.Directory{}, err
	}

	return dto.Directory{
		ID:        dir.ID,
		ChannelID: dir.ChannelID,
		Name:      dir.Name,
		UpdatedAt: now,
	}, nil
}

func (r *RepositoryMessenger) DeleteDirectory(ctx context.Context, dirID uuid.UUID) error {
	deleteDirectory := r.Session.ContextQuery(
		ctx,
		`DELETE FROM messenger_keyspace.directory WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{":id": dirID})

	if err := deleteDirectory.ExecRelease(); err != nil {
		r.Log.Error("error when delete directory: ", "error", err)
		return err
	}

	return nil
}

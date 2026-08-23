package repository

import (
	"context"
	"errors"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

var ErrBotCommandNotFound = errors.New("bot command not found")

func (r *RepositoryBots) InsertBotCommand(ctx context.Context, command dto.BotComand) error {
	insertQuery := `
		INSERT INTO bot_commands (id, bot_id, command, description)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := r.DB.Exec(ctx, insertQuery, command.ID, command.BotID, command.Command, command.Description); err != nil {
		r.Log.Error("error with inserting command: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryBots) SelectBotCommands(ctx context.Context, botID uuid.UUID) ([]dto.BotComand, error) {
	selectBotCommands := `
		SELECT * FROM bot_commands WHERE bot_id=$1
	`

	rows, err := r.DB.Query(ctx, selectBotCommands, botID)
	if err != nil {
		r.Log.Error("error with selecting bot commands: ", "error", err)
		return nil, err
	}
	defer rows.Close()

	var commands []dto.BotComand

	for rows.Next() {
		var command dto.BotComand
		if err := rows.Scan(&command.ID, &command.BotID, &command.Command, &command.Description); err != nil {
			r.Log.Error("error with scaning bot commands: ", "error", err)
			return nil, err
		}

		commands = append(commands, command)
	}

	if err := rows.Err(); err != nil {
		r.Log.Error("error after rows iteration: ", "error", err)
		return nil, err
	}

	return commands, nil
}

func (r *RepositoryBots) UpdateBotCommand(ctx context.Context, command dto.BotComand) error {
	updateQuery := `
		UPDATE bot_commands
		SET command=$3, description=$4
		WHERE id=$1 AND bot_id=$2
	`

	if _, err := r.DB.Exec(ctx, updateQuery, command.ID, command.BotID, command.Command, command.Description); err != nil {
		r.Log.Error("error with updating bot command: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryBots) DeleteBotCommand(ctx context.Context, commandID uuid.UUID) error {
	deleteQuery := `
		DELETE FROM bot_commands WHERE id=$1
	`

	result, err := r.DB.Exec(ctx, deleteQuery, commandID)
	if err != nil {
		r.Log.Error("error with deleting bot command: ", "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrBotCommandNotFound
	}

	return nil
}

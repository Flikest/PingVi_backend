package repository

import (
	"context"
	"errors"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrTokenAlreadyExists = errors.New("bot token already exists")
var ErrBotNotFound = errors.New("bot not found")

func (r *RepositoryBots) InsertBot(ctx context.Context, bot dto.Bot) error {
	insertQuery := `
		INSERT INTO bots (id, name, description, creator_id, token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.DB.Exec(
		ctx,
		insertQuery,
		bot.ID,
		bot.Name,
		bot.Description,
		bot.Avatar,
		bot.CreatorID,
		bot.Token,
		bot.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				r.Log.Warn("attempt to insert bot with duplicate token", "token", bot.Token)
				return ErrTokenAlreadyExists
			}
		}

		r.Log.Error("error with inserting bot: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryBots) SelectBotByID(ctx context.Context, botID uuid.UUID) (dto.BotResponse, error) {
	selectQuery := `
		SELECT 
		b.id, b.name, b.description, b.avatar, b.creator_id, b.token, b.created_at,
		c.id, c.bot_id, c.command, c.description
		FROM bots b
		LEFT JOIN bot_commands c ON b.id = c.bot_id 
		WHERE b.id = $1;
	`

	rows, err := r.DB.Query(ctx, selectQuery, botID)
	if err != nil {
		r.Log.Error("error with query bot: ", "error", err)
		return dto.BotResponse{}, err
	}
	defer rows.Close()

	var response dto.BotResponse
	var botFilled bool

	for rows.Next() {
		var cmdID, cmdBotID *uuid.UUID
		var cmdName, cmdDesc *string

		err := rows.Scan(
			&response.Bot.ID,
			&response.Bot.Name,
			&response.Bot.Description,
			&response.Bot.Avatar,
			&response.Bot.CreatorID,
			&response.Bot.Token,
			&response.Bot.CreatedAt,
			&cmdID,
			&cmdBotID,
			&cmdName,
			&cmdDesc,
		)
		if err != nil {
			r.Log.Error("error scanning bot row: ", "error", err)
			return dto.BotResponse{}, err
		}

		botFilled = true

		if cmdID != nil {
			command := dto.BotComand{
				ID:          *cmdID,
				BotID:       *cmdBotID,
				Command:     *cmdName,
				Description: *cmdDesc,
			}
			response.Commands = append(response.Commands, command)
		}
	}

	if err := rows.Err(); err != nil {
		return dto.BotResponse{}, err
	}

	if !botFilled {
		return dto.BotResponse{}, pgx.ErrNoRows
	}

	return response, nil
}

func (r *RepositoryBots) SelectBotsByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]dto.Bot, error) {
	selectBotsQuery := `
		SELECT * 
		FROM bots 
		WHERE creator_id = $1
	`

	var bots []dto.Bot

	rows, err := r.DB.Query(ctx, selectBotsQuery)
	if err != nil {
		r.Log.Error("error with selecting rows by creator id: ", "error", err)
		return nil, err
	}

	for rows.Next() {
		var bot dto.Bot
		if err := rows.Scan(
			&bot.ID,
			&bot.Name,
			&bot.Description,
			&bot.CreatorID,
			&bot.Token,
			&bot.CreatedAt,
		); err != nil {
			r.Log.Error("error with scaning bot: ", "error", err)
			return nil, err
		}

		bots = append(bots, bot)
	}

	if err := rows.Err(); err != nil {
		r.Log.Error("error during rows iteration", "error", err)
		return nil, err
	}

	return bots, nil
}

func (r *RepositoryBots) UpdateBot(ctx context.Context, bot dto.Bot) error {
	updateQuery := `
		UPDATE bots
		SET name=$3, description=$4 avatar=$5
		WHERE id=$1, creator_id=$2
	`

	if _, err := r.DB.Exec(ctx, updateQuery, bot.ID, bot.CreatorID, bot.Name, bot.Description, bot.Avatar); err != nil {
		r.Log.Error("error with updating bot: ", "error", err)
		return err
	}

	return nil
}

func (r *RepositoryBots) DeleteBot(ctx context.Context, botID uuid.UUID) error {
	deleteQuery := `
		DELETE FROM bots WHERE id=$1
	`

	result, err := r.DB.Exec(ctx, deleteQuery, botID)
	if err != nil {
		r.Log.Error("error with deleting bot: ", "error", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrBotNotFound
	}

	return nil
}

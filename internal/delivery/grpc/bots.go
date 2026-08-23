package deliverygrpc

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	pb "github.com/Flikest/PingVi_backend/gen/go/bots"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	servicegrpc "github.com/Flikest/PingVi_backend/internal/services/grpc"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type HandlerBots struct {
	pb.UnimplementedBotsServer
	Service *servicegrpc.ServiceBots
	Hub     *servicehttp.BotHub
	Log     *slog.Logger

	mu sync.RWMutex
}

func (h *HandlerBots) GetBotAnswers(
	stream pb.Bots_GetBotAnswersServer,
) error {
	ctx := stream.Context()
	errChan := make(chan error, 2)

	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}

			ID, err := snowflake.ParseString(req.MessageData.Message.Id)
			if err != nil {
				h.Log.Error("error with parsing snowflake message id", "error", err)
				continue
			}

			chatID, err := uuid.Parse(req.MessageData.Message.ChatId)
			if err != nil {
				h.Log.Error("error with parsing uuid chat id", "error", err)
				continue
			}

			senderID, err := uuid.Parse(req.MessageData.Message.SenderId)
			if err != nil {
				h.Log.Error("error with parsing uuid sender id", "error", err)
				continue
			}

			replyToID, err := snowflake.ParseString(req.MessageData.Message.ReplyToId)
			if err != nil {
				h.Log.Error("error with parsing snowflake reply to id", "error", err)
				continue
			}

			msCreatedAt, err := strconv.ParseInt(req.MessageData.Message.CreatedAt, 10, 64)
			if err != nil {
				h.Log.Error("error with parsing time created at", "error", err)
				continue
			}

			msUpdatedAt, err := strconv.ParseInt(req.MessageData.Message.UpdatedAt, 10, 64)
			if err != nil {
				h.Log.Error("error with parsing time created at", "error", err)
				continue
			}

			createdAt := time.UnixMilli(msCreatedAt)
			updatedAt := time.UnixMilli(msUpdatedAt)

			message := dto.Message{
				ID:        ID,
				ChatID:    chatID,
				SenderID:  senderID,
				Message:   req.MessageData.Message.Message,
				IsMy:      req.MessageData.Message.IsMy,
				IsSystem:  req.MessageData.Message.IsSystem,
				ReplyToID: replyToID,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			reactionID, err := uuid.Parse(req.MessageData.Reaction.Id)
			if err != nil {
				h.Log.Error("error with parsing uuid reaction id", "error", err)
				continue
			}

			sendedAt := time.UnixMilli(msCreatedAt)

			reaction := dto.Reaction{
				ID:        reactionID,
				ChatID:    chatID,
				UserID:    senderID,
				MessageID: ID,
				Reaction:  req.MessageData.Reaction.Reaction,
				SendedAt:  sendedAt,
			}

			messageNew := servicehttp.Message{
				Operation: req.MessageData.Operation,
				Message:   message,
				Reaction:  reaction,
				Status:    req.MessageData.Status,
			}

			h.mu.Lock()
			ch, exists := h.Hub.Messages[req.BotID]
			if !exists {
				ch = make(chan []servicehttp.Message, 100)
				h.Hub.Messages[req.BotID] = ch
			}
			h.mu.Unlock()

			select {
			case ch <- []servicehttp.Message{messageNew}:
			default:
				h.Log.Error("buffer overflow for bot channel, message dropped", "botID", req.BotID)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-h.Hub.Answers:
				if !ok {
					return
				}

				messageNew := &pb.Message{
					Operation: message.Operation,
					Message: &pb.MessageItem{
						Id:        message.Message.ID.String(),
						ChatId:    message.Message.ChatID.String(),
						SenderId:  message.Message.SenderID.String(),
						Message:   message.Message.Message,
						IsMy:      message.Message.IsMy,
						IsSystem:  message.Message.IsSystem,
						ReplyToId: message.Message.ReplyToID.String(),
						CreatedAt: message.Message.CreatedAt.String(),
						UpdatedAt: message.Message.UpdatedAt.String(),
					},
					Reaction: &pb.Reaction{
						Id:        message.Reaction.ID.String(),
						ChatId:    message.Reaction.ChatID.String(),
						UserId:    message.Reaction.UserID.String(),
						MessageId: message.Reaction.MessageID.String(),
						Reaction:  message.Reaction.Reaction,
						SendedAt:  message.Reaction.SendedAt.String(),
					},
					Status: message.Status,
				}

				if err := stream.Send(&pb.StreamMessageResponse{MessageData: messageNew}); err != nil {
					h.Log.Error("error sending stream response", "error", err)
					errChan <- err
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (h *HandlerBots) GetBotInfo(
	ctx context.Context,
	in *pb.GetBotInfoRequest,
) (*pb.GetBotInfoResponse, error) {
	if in.GetBotId() == "" {
		h.Log.Warn("Bot id is a required field")
		return nil, errors.New("bot ID is a required field")
	}

	bot, err := h.Service.GetBotInfo(ctx, in.GetBotId())
	if err != nil {
		h.Log.Error("error with getting bot by id: ", "error", err)
		return nil, err
	}

	return &pb.GetBotInfoResponse{
		Id:          bot.Bot.ID.String(),
		Name:        bot.Bot.Name,
		Description: bot.Bot.Description,
		Avatar:      bot.Bot.Avatar,
		CreatorId:   bot.Bot.CreatorID.String(),
		Token:       bot.Bot.Token,
		CreatedAt:   bot.Bot.CreatedAt.String(),
	}, nil
}

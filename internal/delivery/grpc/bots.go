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
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				req, err := stream.Recv()
				if err != nil {
					select {
					case errChan <- err:
					case <-ctx.Done():
					}
					return
				}

				ID, err := snowflake.ParseString(req.MessageData.Message.Id)
				if err != nil {
					h.Log.Error("error with parsing snowflake message id", "error", err, "id", req.MessageData.Message.Id)
					continue
				}

				chatID, err := uuid.Parse(req.MessageData.Message.ChatId)
				if err != nil {
					h.Log.Error("error with parsing uuid chat id", "error", err, "chatId", req.MessageData.Message.ChatId)
					continue
				}

				senderID, err := uuid.Parse(req.MessageData.Message.SenderId)
				if err != nil {
					h.Log.Error("error with parsing uuid sender id", "error", err, "senderId", req.MessageData.Message.SenderId)
					continue
				}

				var replyToID snowflake.ID
				if req.MessageData.Message.ReplyToId != "" {
					replyToID, err = snowflake.ParseString(req.MessageData.Message.ReplyToId)
					if err != nil {
						h.Log.Error("error with parsing snowflake reply to id", "error", err, "replyToId", req.MessageData.Message.ReplyToId)
						continue
					}
				}

				msCreatedAt, err := strconv.ParseInt(req.MessageData.Message.CreatedAt, 10, 64)
				if err != nil {
					h.Log.Error("error with parsing time created at", "error", err, "createdAt", req.MessageData.Message.CreatedAt)
					continue
				}

				msUpdatedAt, err := strconv.ParseInt(req.MessageData.Message.UpdatedAt, 10, 64)
				if err != nil {
					h.Log.Error("error with parsing time updated at", "error", err, "updatedAt", req.MessageData.Message.UpdatedAt)
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
					h.Log.Error("error with parsing uuid reaction id", "error", err, "reactionId", req.MessageData.Reaction.Id)
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
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-h.Hub.Answers:
				if !ok {
					h.Log.Warn("answers channel closed")
					return
				}

				if len(message) == 0 {
					continue
				}

				for i := range message {
					if message[i].Message.ID == 0 {
						h.Log.Warn("skipping message with zero ID")
						continue
					}

					msg := &pb.Message{
						Operation: message[i].Operation,
						Message: &pb.MessageItem{
							Id:        message[i].Message.ID.String(),
							ChatId:    message[i].Message.ChatID.String(),
							SenderId:  message[i].Message.SenderID.String(),
							Message:   message[i].Message.Message,
							IsMy:      message[i].Message.IsMy,
							IsSystem:  message[i].Message.IsSystem,
							ReplyToId: message[i].Message.ReplyToID.String(),
							CreatedAt: message[i].Message.CreatedAt.String(),
							UpdatedAt: message[i].Message.UpdatedAt.String(),
						},
						Reaction: &pb.Reaction{
							Id:        message[i].Reaction.ID.String(),
							ChatId:    message[i].Reaction.ChatID.String(),
							UserId:    message[i].Reaction.UserID.String(),
							MessageId: message[i].Reaction.MessageID.String(),
							Reaction:  message[i].Reaction.Reaction,
							SendedAt:  message[i].Reaction.SendedAt.String(),
						},
						Status: message[i].Status,
					}

					select {
					case <-ctx.Done():
						h.Log.Info("context cancelled, stopping send")
						return
					default:
						if err := stream.Send(&pb.StreamMessageResponse{
							MessageData: msg,
						}); err != nil {
							h.Log.Error("error sending stream response", "error", err)
							select {
							case errChan <- err:
							case <-ctx.Done():
							}
							return
						}
					}
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		h.Log.Info("context done", "error", ctx.Err())
		return ctx.Err()
	case err := <-errChan:
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return err
	case <-done:
		h.Log.Info("receive goroutine finished")
		return nil
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

package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s ServiceMessenger) GetStartMyChats(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channels, err := s.Repository.SelectChannels(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select channels from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groups, err := s.Repository.SelectGroup(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select groups from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	personalChats, err := s.Repository.SelectPersonalChats(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select personal chats from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var communities []dto.StartСommunitiesData

	for i := 0; i < max(len(groups), len(channels), len(personalChats)); i++ {

		if len(groups) > 0 {
			lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), groups[i].ID)
			if err != nil {
				s.Log.Error("error with select last message by chat id: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			unreadMesages, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, groups[i].ID)
			if err != nil {
				s.Log.Error("error with select unread messages count: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			isMuted, err := s.Repository.IsGroupMuted(ctx.Request.Context(), groups[i].ID, userID)
			if err != nil {
				s.Log.Error("error with select muted in group members: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			communities = append(communities, dto.StartСommunitiesData{
				Type:           "group",
				Communities:    groups[i],
				LastMessage:    lastMessage.Message,
				IsBot:          false,
				IsMuted:        isMuted,
				UnreadMessages: int(unreadMesages),
			})
			groups = groups[1:]
		}

		if len(channels) > 0 {
			lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), channels[i].ID)
			if err != nil {
				s.Log.Error("error with select last message by chat id: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			unreadMesages, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, channels[i].ID)
			if err != nil {
				s.Log.Error("error with select unread messages count: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			isMuted, err := s.Repository.IsChannelMuted(ctx.Request.Context(), channels[i].ID, userID)
			if err != nil {
				s.Log.Error("error with select muted in channel members: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			communities = append(communities, dto.StartСommunitiesData{
				Type:           "channel",
				Communities:    channels[i],
				LastMessage:    lastMessage.Message,
				IsBot:          false,
				IsMuted:        isMuted,
				UnreadMessages: int(unreadMesages),
			})
			channels = channels[1:]
		}

		if len(personalChats) > 0 {
			lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), personalChats[i].ID)
			if err != nil {
				s.Log.Error("error with select last message by chat id: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			unreadMesages, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, personalChats[i].ID)
			if err != nil {
				s.Log.Error("error with select unread messages count: ", "error", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			communities = append(communities, dto.StartСommunitiesData{
				Type:           "personal",
				Communities:    personalChats[i],
				LastMessage:    lastMessage.Message,
				IsBot:          false,
				IsMuted:        false,
				UnreadMessages: int(unreadMesages),
			})
			personalChats = personalChats[1:]
		}
	}

	ctx.JSON(http.StatusOK, communities)
}

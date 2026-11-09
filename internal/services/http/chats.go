package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

func (s ServiceMessenger) GetStartMyChats(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := gocql.ParseUUID(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groups, err := s.Repository.SelectGroup(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select groups from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	channels, err := s.Repository.SelectChannels(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select channels from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	personalChats, err := s.Repository.SelectPersonalChats(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("error with select personal chats from db: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	communities := make([]dto.StartСommunitiesData, 0, len(groups)+len(channels)+len(personalChats))

	// ---------- groups ----------
	for _, group := range groups {
		lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), group.ID)
		if err != nil {
			s.Log.Error("error with select last message by chat id: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		unread, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, group.ID)
		if err != nil {
			s.Log.Error("error with select unread messages count: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		isMuted, err := s.Repository.IsGroupMuted(ctx.Request.Context(), group.ID, userID)
		if err != nil {
			s.Log.Error("error with select muted in group members: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lastMsgText := ""
		if lastMessage != nil {
			lastMsgText = lastMessage.Message
		}

		communities = append(communities, dto.StartСommunitiesData{
			Type:           "group",
			Communities:    group,
			LastMessage:    lastMsgText,
			IsBot:          false,
			IsMuted:        isMuted,
			UnreadMessages: int(unread),
		})
	}

	// ---------- channels ----------
	for _, channel := range channels {
		lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), channel.ID)
		if err != nil {
			s.Log.Error("error with select last message by chat id: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		unread, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, channel.ID)
		if err != nil {
			s.Log.Error("error with select unread messages count: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		isMuted, err := s.Repository.IsChannelMuted(ctx.Request.Context(), channel.ID, userID)
		if err != nil {
			s.Log.Error("error with select muted in channel members: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lastMsgText := ""
		if lastMessage != nil {
			lastMsgText = lastMessage.Message
		}

		communities = append(communities, dto.StartСommunitiesData{
			Type:           "channel",
			Communities:    channel,
			LastMessage:    lastMsgText,
			IsBot:          false,
			IsMuted:        isMuted,
			UnreadMessages: int(unread),
		})
	}

	// ---------- personal chats ----------
	// Запрашиваем всех собеседников ОДНИМ gRPC-вызовом
	// и раскладываем нужные поля в мапу (не завязываемся на тип user_info.*).
	type userInfo struct {
		Name   string
		Avatar string
	}
	userMap := make(map[string]userInfo, len(personalChats))

	if len(personalChats) > 0 {
		ids := make([]string, 0, len(personalChats))
		for _, pc := range personalChats {
			ids = append(ids, pc.UserID.String())
		}

		resp, err := s.Client.GetUsersByIds(ctx.Request.Context(), &user_info.GetUserByIDRequest{UserIds: ids})
		if err != nil {
			s.Log.Error("error with getting users by ids: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, u := range resp.Users {
			avatar := ""
			if len(u.Avatar) > 0 {
				avatar = u.Avatar[0]
			}
			userMap[u.Id] = userInfo{Name: u.Name, Avatar: avatar}
		}
	}

	for _, pc := range personalChats {
		lastMessage, err := s.Repository.SelectLastMessageByChatID(ctx.Request.Context(), pc.ID)
		if err != nil {
			s.Log.Error("error with select last message by chat id: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		unread, err := s.Repository.SelectUnreadCountMessage(ctx.Request.Context(), userID, pc.ID)
		if err != nil {
			s.Log.Error("error with select unread messages count: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if u, ok := userMap[pc.UserID.String()]; ok {
			pc.Name = u.Name
			pc.Avatar = u.Avatar
		}

		lastMsgText := ""
		if lastMessage != nil {
			lastMsgText = lastMessage.Message
		}

		communities = append(communities, dto.StartСommunitiesData{
			Type:           "personal_chat",
			Communities:    pc,
			LastMessage:    lastMsgText,
			IsBot:          false,
			IsMuted:        false,
			UnreadMessages: int(unread),
		})
	}

	ctx.JSON(http.StatusOK, communities)
}

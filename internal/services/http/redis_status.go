package servicehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	KeyUserOnline     = "user:online:%s"
	KeyUserActiveChat = "user:active:%s"
	KeyUserTyping     = "user:typing:%s"
	KeyUserSubs       = "user:subs:%s"

	OnlineTTL     = 30 * time.Second
	TypingTTL     = 10 * time.Second
	ActiveChatTTL = 5 * time.Minute
)

func (h *MessengerHub) publishUserStatus(ctx context.Context, userID gocql.UUID) error {
	status := map[string]interface{}{
		"online":      "online",
		"last_seen":   time.Now().Unix(),
		"active_chat": "none",
		"typing_chat": "none",
		"is_typing":   false,
	}

	activeChat, err := h.getUserActiveChat(ctx, userID)
	if err != nil {
		return err
	}
	if activeChat != gocql.UUID(uuid.Nil) {
		status["active_chat"] = activeChat
	}

	typingChat, err := h.getUserTyping(ctx, userID)
	if err != nil {
		return err
	}
	if typingChat != gocql.UUID(uuid.Nil) {
		status["typing_chat"] = typingChat
		status["is_typing"] = true
	}

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(KeyUserOnline, userID)
	return h.Redis.Publish(ctx, key, data).Err()
}

func (h *MessengerHub) getSubscribedUsersStatus(ctx context.Context, userID gocql.UUID) (map[gocql.UUID]map[string]interface{}, error) {
	subs, err := h.getUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	statuses := make(map[gocql.UUID]map[string]interface{})
	for _, subID := range subs {
		status := map[string]interface{}{
			"online":      "online",
			"last_seen":   time.Now().Unix(),
			"active_chat": "none",
			"typing_chat": "none",
			"is_typing":   false,
		}

		online, err := h.isUserOnline(ctx, subID)
		if err != nil {
			h.Log.Warn("failed to check online status", "user_id", subID, "error", err)
		}
		if online {
			status["online"] = "online"

			activeChat, err := h.getUserActiveChat(ctx, subID)
			if err != nil {
				h.Log.Warn("failed to get active chat", "user_id", subID, "error", err)
			} else if activeChat != gocql.UUID(uuid.Nil) {
				status["active_chat"] = activeChat
			}

			typingChat, err := h.getUserTyping(ctx, subID)
			if err != nil {
				h.Log.Warn("failed to get typing status", "user_id", subID, "error", err)
			} else if typingChat != gocql.UUID(uuid.Nil) {
				status["typing_chat"] = typingChat
				status["is_typing"] = true
			}
		}

		statuses[subID] = status
	}

	return statuses, nil
}

func (h *MessengerHub) getSubscribedUsersPresences(ctx context.Context, userID gocql.UUID) ([]dto.UserPresence, error) {
	subs, err := h.getUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	presences := make([]dto.UserPresence, 0, len(subs))
	for _, subID := range subs {
		activeChat, err := h.getUserActiveChat(ctx, subID)
		if err != nil {
			h.Log.Warn("failed to get active chat", "user_id", subID, "error", err)
			activeChat = gocql.UUID(uuid.Nil)
		}

		typingChat, err := h.getUserTyping(ctx, subID)
		if err != nil {
			h.Log.Warn("failed to get typing status", "user_id", subID, "error", err)
			typingChat = gocql.UUID(uuid.Nil)
		}

		online, err := h.isUserOnline(ctx, subID)
		if err != nil {
			h.Log.Warn("failed to check online status", "user_id", subID, "error", err)
			online = false
		}

		presences = append(presences, dto.UserPresence{
			UserID:     subID,
			ActiveChat: activeChat,
			TypingChat: typingChat,
			IsOnline:   online,
			LastSeen:   time.Now(),
		})
	}

	return presences, nil
}

func (h *MessengerHub) getUserActiveChat(ctx context.Context, userID gocql.UUID) (gocql.UUID, error) {
	key := fmt.Sprintf(KeyUserActiveChat, userID)
	chatIDStr, err := h.Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return gocql.UUID{}, nil
		}
		return gocql.UUID{}, err
	}
	return gocql.ParseUUID(chatIDStr)
}

func (h *MessengerHub) getUserTyping(ctx context.Context, userID gocql.UUID) (gocql.UUID, error) {
	key := fmt.Sprintf(KeyUserTyping, userID)
	chatIDStr, err := h.Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return gocql.UUID{}, nil
		}
		return gocql.UUID{}, err
	}
	return gocql.ParseUUID(chatIDStr)
}

func (h *MessengerHub) getUserSubscriptions(ctx context.Context, userID gocql.UUID) ([]gocql.UUID, error) {
	key := fmt.Sprintf(KeyUserSubs, userID)
	members, err := h.Redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var userIDs []gocql.UUID
	for _, member := range members {
		if userID, err := gocql.ParseUUID(member); err == nil {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs, nil
}

func (h *MessengerHub) isUserOnline(ctx context.Context, userID gocql.UUID) (bool, error) {
	key := fmt.Sprintf(KeyUserOnline, userID)
	exists, err := h.Redis.Exists(ctx, key).Result()
	return exists == 1, err
}

func (h *MessengerHub) setUserTyping(ctx context.Context, userID, chatID gocql.UUID) error {
	key := fmt.Sprintf(KeyUserTyping, userID)
	return h.Redis.Set(ctx, key, chatID.String(), TypingTTL).Err()
}

func (h *MessengerHub) clearUserTyping(ctx context.Context, userID gocql.UUID) error {
	key := fmt.Sprintf(KeyUserTyping, userID)
	return h.Redis.Del(ctx, key).Err()
}

func (h *MessengerHub) notifySubscribers(ctx context.Context, userID gocql.UUID, statusType string, chatID gocql.UUID) {
	if h.Redis == nil {
		return
	}

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.Clients))
	for _, client := range h.Clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		clientUserID, err := gocql.ParseUUID(client.ID)
		if err != nil {
			continue
		}

		subs, err := h.getUserSubscriptions(ctx, clientUserID)
		if err != nil {
			continue
		}

		for _, subID := range subs {
			if subID == userID {
				statusMessage := Message{
					Operation: "user_status_update",
					Message: dto.Message{
						SenderID: userID,
						ChatID:   chatID,
					},
				}

				if statusType == "typing" {
					statusMessage.Status = "typing_start"
					statusMessage.Message.ChatID = chatID
				} else if statusType == "active_chat" {
					statusMessage.Status = "active_chat_changed"
				}

				client.Push(statusMessage)
				break
			}
		}
	}
}

package servicehttp

import (
	"context"
	"fmt"
	"time"

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

func (h *Hub) setUserOnline(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserOnline, userID)
	return h.Redis.Set(ctx, key, time.Now().Unix(), OnlineTTL).Err()
}

func (h *Hub) setUserOffline(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserOnline, userID)
	return h.Redis.Del(ctx, key).Err()
}

func (h *Hub) isUserOnline(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := fmt.Sprintf(KeyUserOnline, userID)
	exists, err := h.Redis.Exists(ctx, key).Result()
	return exists == 1, err
}

func (h *Hub) setUserActiveChat(ctx context.Context, userID, chatID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserActiveChat, userID)
	return h.Redis.Set(ctx, key, chatID.String(), ActiveChatTTL).Err()
}

func (h *Hub) getUserActiveChat(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	key := fmt.Sprintf(KeyUserActiveChat, userID)
	chatIDStr, err := h.Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return uuid.Parse(chatIDStr)
}

func (h *Hub) setUserTyping(ctx context.Context, userID, chatID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserTyping, userID)
	return h.Redis.Set(ctx, key, chatID.String(), TypingTTL).Err()
}

func (h *Hub) clearUserTyping(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserTyping, userID)
	return h.Redis.Del(ctx, key).Err()
}

func (h *Hub) getUserTyping(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	key := fmt.Sprintf(KeyUserTyping, userID)
	chatIDStr, err := h.Redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return uuid.Parse(chatIDStr)
}

func (h *Hub) addUserSubscription(ctx context.Context, userID, subUserID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserSubs, userID)
	return h.Redis.SAdd(ctx, key, subUserID.String()).Err()
}

func (h *Hub) getUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	key := fmt.Sprintf(KeyUserSubs, userID)
	members, err := h.Redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var userIDs []uuid.UUID
	for _, member := range members {
		if userID, err := uuid.Parse(member); err == nil {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs, nil
}

func (h *Hub) removeUserSubscription(ctx context.Context, userID, subUserID uuid.UUID) error {
	key := fmt.Sprintf(KeyUserSubs, userID)
	return h.Redis.SRem(ctx, key, subUserID.String()).Err()
}

func (h *Hub) getSubscribedUsersStatus(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]map[string]interface{}, error) {
	subs, err := h.getUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	statuses := make(map[uuid.UUID]map[string]interface{})
	for _, subID := range subs {
		status := make(map[string]interface{})

		online, _ := h.isUserOnline(ctx, subID)
		status["online"] = online

		activeChat, _ := h.getUserActiveChat(ctx, subID)
		if activeChat != uuid.Nil {
			status["active_chat"] = activeChat
		}

		typingChat, _ := h.getUserTyping(ctx, subID)
		if typingChat != uuid.Nil {
			status["typing_chat"] = typingChat
			status["is_typing"] = true
		}

		statuses[subID] = status
	}

	return statuses, nil
}

package redis

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	ctx  context.Context
	conn string
}

func NewRedisCleint(rc *RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return rdb
}

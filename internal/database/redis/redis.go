package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisCleint() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return rdb
}

package services

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService() *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return &RedisService{client: client}
}

func (r *RedisService) Close() {
	r.client.Close()
}

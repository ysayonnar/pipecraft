package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"pipecraft/internal/logger"
	"time"
)

const DEFAULT_TTL_SECONDS = 30

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

func (r *RedisService) setValueByKey(key string, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := r.client.Set(ctx, key, value, time.Second*time.Duration(DEFAULT_TTL_SECONDS)).Err()
	if err != nil {
		slog.Error("error while setting data to redis", logger.Err(err))
	}
}

func (r *RedisService) SetPipelineStatus(id int64, data string) {
	r.setValueByKey(fmt.Sprintf("status:%d", id), data)
}

func (r *RedisService) SetPipelineLogs(id int64, data string) {
	r.setValueByKey(fmt.Sprintf("logs:%d", id), data)
}

func (r *RedisService) getStringByKey(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			slog.Error("error while getting data from redis", logger.Err(err))
		}
		return ""
	}

	return data
}

func (r *RedisService) GetPipelineStatus(id int64) string {
	return r.getStringByKey(fmt.Sprintf("status:%d", id))
}

func (r *RedisService) GetPipelineLogs(id int64) string {
	return r.getStringByKey(fmt.Sprintf("logs:%d", id))
}

package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, pass string) *RedisCache {
	return &RedisCache{
		redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: pass,
			DB:       0,
		}),
	}
}

func (r *RedisCache) Ping(ctx context.Context) {
	pong, err := r.client.Ping(ctx).Result()
	if err != nil {
		slog.Error("redis ping", "error", err)
	} else {
		slog.Info("redis ping", "response", pong)
	}
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()

	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	_, err := r.client.Del(ctx, keys...).Result()

	if err != nil {
		return err
	}

	return nil
}

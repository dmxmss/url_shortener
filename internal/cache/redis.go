package cache

import (
	"context"
	"errors"
	"time"

	"dmxmss-project/internal/config"
	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("cache miss")

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context, cfg config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

func (r *Redis) GetURL(ctx context.Context, code string) (string, error) {
	value, err := r.client.Get(ctx, "url:"+code).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return value, err
}

func (r *Redis) SetURL(ctx context.Context, code, longURL string, ttl time.Duration) error {
	return r.client.Set(ctx, "url:"+code, longURL, ttl).Err()
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}

package redisutils

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	Ping(ctx context.Context) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration int) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Lpush(ctx context.Context, key string, values ...interface{}) (int64, error)
	Rpop(ctx context.Context, key string) (string, error)
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context) *redisClient {
	options := &redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "", // no password set
		DB:       0,
	}
	c := redis.NewClient(options)
	return &redisClient{client: c}
}

func (r *redisClient) Ping(ctx context.Context) (string, error) {
	return r.client.Ping(ctx).Result()
}

func (r *redisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.HSet(ctx, key, values...).Result()
}

func (r *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration int) (string, error) {
	return r.client.Set(ctx, key, value, 0).Result()
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Lpush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.LPush(ctx, key, values...).Result()
}

func (r *redisClient) Rpop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

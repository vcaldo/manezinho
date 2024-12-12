package redisutils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	RedisDownloadsKey = "completed_downloads"
)

type Download struct {
	ID         int64
	Name       string
	Path       string
	UploadPath string
}

func NewRedisClient(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return rdb, nil
}

func NewAuthenticatedRedisClient(ctx context.Context) (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	password := ""
	db := 0
	return NewRedisClient(ctx, addr, password, db)
}

func DownloadExistsInRedis(ctx context.Context, rdb *redis.Client, id int64) (bool, error) {
	val, err := rdb.HExists(ctx, RedisDownloadsKey, strconv.FormatInt(id, 10)).Result()
	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}
	return val, nil
}

func RegisterDownloadState(ctx context.Context, rdb *redis.Client, d Download) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	log.Printf("Storing download in Redis: %s", d.Name)
	err = rdb.HSet(ctx, RedisDownloadsKey, strconv.FormatInt(d.ID, 10), string(data)).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

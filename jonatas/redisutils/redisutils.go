package redisutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	CompletedHash = "completed"
)

type Download struct {
	ID         int64
	Name       string
	Path       string
	UploadPath string
	State      string
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
	val, err := rdb.HExists(ctx, fmt.Sprintf("%s:%d", CompletedHash, id), "state").Result()
	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}

	return val, nil
}

func RegisterDownloadState(ctx context.Context, rdb *redis.Client, d Download) error {
	log.Printf("Storing download in Redis: %s", d.Name)
	err := rdb.HSet(ctx, fmt.Sprintf("%s:%d", CompletedHash, d.ID), []string{NameKey, d.Name, StateKey, d.State}).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	switch d.State {
	case Downloaded:
		err = rdb.SAdd(ctx, Downloaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Compressing:
		err = rdb.SAdd(ctx, Compressing, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err := rdb.SRem(ctx, Downloaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Compressed:
		err = rdb.SAdd(ctx, Compressed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = rdb.SRem(ctx, Compressing, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Uploading:
		err = rdb.SAdd(ctx, Uploading, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = rdb.SRem(ctx, Compressed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Uploaded:
		err = rdb.SAdd(ctx, Uploaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = rdb.SRem(ctx, Uploading, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	case Removed:
		err = rdb.SAdd(ctx, Removed, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
		err = rdb.SRem(ctx, Uploaded, d.ID).Err()
		if err != nil {
			return fmt.Errorf("redis set failed: %w", err)
		}
	default:
		log.Println("Unknown state")
	}

	return nil
}

func GetDowloadState(ctx context.Context, rdb *redis.Client, state string) ([]int64, error) {
	val, err := rdb.SMembers(ctx, state).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	// Convert string slice to int64 slice
	var ids []int64
	for _, v := range val {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

package utils

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/hekmon/transmissionrpc"
	"github.com/vcaldo/manezinho/bot/transmission"
)

func newTransmissionClient(ctx context.Context) (*transmission.Client, error) {
	url := os.Getenv("TRANSMISSION_URL")
	user := os.Getenv("TRANSMISSION_USER")
	pass := os.Getenv("TRANSMISSION_PASS")
	return transmission.NewClient(ctx, url, user, pass)
}

func AddTorrentFromFile(ctx context.Context, b *bot.Bot, fileID string, fileName string) (*transmissionrpc.Torrent, error) {
	c, err := newTransmissionClient(ctx)
	if err != nil {
		log.Printf("failed to create transmission client: %v", err)
		return nil, err
	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		log.Printf("failed to get file: %v", err)
		return nil, err
	}

	addedTorrent, err := c.AddTorrentFromFile(ctx, file.FilePath)
	if err != nil {
		log.Printf("failed to add torrent: %v", err)
		return nil, err
	}

	return addedTorrent, nil
}

func AddTorrentFromMagnet(ctx context.Context, msg string) (*transmissionrpc.Torrent, error) {
	c, err := newTransmissionClient(ctx)
	if err != nil {
		log.Printf("failed to create transmission client: %v", err)
		return nil, err
	}

	addedTorrent, err := c.AddTorrent(ctx, msg)
	if err != nil {
		return nil, err
	}

	return addedTorrent, nil
}

func IsUserAllowed(ctx context.Context, userId int64) bool {
	// Create a slice of allowed user ids from env var
	allowedUserIds := os.Getenv("ALLOWED_USER_IDS")
	allowedUserIdsSlice := strings.Split(allowedUserIds, ",")
	allowedUserIdsInt64 := make([]int64, len(allowedUserIdsSlice))
	for i, id := range allowedUserIdsSlice {
		var err error
		allowedUserIdsInt64[i], err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Printf("failed to parse user id: %v", err)
			return false
		}
	}

	// Check if the user id is in the allowed user ids slice
	for _, id := range allowedUserIdsInt64 {
		if id == userId {
			log.Printf("user %v is allowed to use the bot", userId)
			return true
		}
	}
	log.Printf("user %v is allowed to use the bot", userId)

	return false
}

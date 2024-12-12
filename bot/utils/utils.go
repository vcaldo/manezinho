package utils

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/hekmon/transmissionrpc"
	"github.com/vcaldo/manezinho/bot/transmission"
)

func AddTorrentFromFile(ctx context.Context, b *bot.Bot, fileID string, fileName string) (*transmissionrpc.Torrent, error) {
	c, err := transmission.NewTransmissionClient(ctx)
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
	c, err := transmission.NewTransmissionClient(ctx)
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

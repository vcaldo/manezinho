package utils

import (
	"context"
	"log"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/manezinho/bot/redisutils"
	"github.com/vcaldo/manezinho/bot/transmission"
)

func GetCompletedDownloads(ctx context.Context, downloadChan chan<- redisutils.Download) error {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
		return err
	}

	rdb, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
	}
	defer rdb.Close()

	completedDownloads, err := c.GetCompletedDownloads(ctx)
	if err != nil {
		log.Printf("error getting completed downloads: %v", err)
		return err
	}

	for _, download := range completedDownloads {
		d := redisutils.Download{
			ID:         *download.ID,
			Name:       *download.Name,
			Path:       filepath.Join(ComplatedDownloadsPath, *download.Name),
			UploadPath: filepath.Join(UploadsReadyPath, *download.Name),
		}
		// Check if download exists in Redis
		exists, err := redisutils.DownloadExistsInRedis(ctx, rdb, d.ID)
		if err != nil {
			log.Printf("error checking redis: %v", err)
			continue
		}

		// Store in Redis and push to channel if new
		if !exists {
			log.Printf("New download completion detected: %s", d.Name)
			if err := redisutils.StoreDownloadInRedis(ctx, rdb, d); err != nil {
				log.Printf("error storing in redis: %v", err)
				continue
			}
			downloadChan <- d
		}
	}
	return nil
}

func ProcessDownload(ctx context.Context, b *bot.Bot, download redisutils.Download) {
	log.Printf("Processing download %s", download.Name)
	// Compress download
	err := CompressDownload(ctx, download)
	if err != nil {
		log.Printf("error compressing and splitting download: %v", err)
		return
	}

	// Upload download
	err = UploadDir(ctx, b, download)
	if err != nil {
		log.Printf("error uploading download: %v", err)
		return
	}

	// Cleanup
	err = RemoveUploadedFiles(ctx, download)
	if err != nil {
		log.Printf("error removing uploaded files: %v", err)
		return
	}

	// Remove Compressed download
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
		return
	}

	// Remove torrent from Transmission
	err = c.RemoveTorrents(ctx, []int64{download.ID})
	if err != nil {
		log.Printf("error removing torrent: %v", err)
		return
	}

	log.Printf("Finished processing download %s", download.Name)
}

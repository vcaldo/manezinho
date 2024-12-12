package utils

import (
	"context"
	"log"
	"path/filepath"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/manezinho/bot/transmission"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

const (
	ComplatedDownloadsPath = "/downloads/complete"
	UploadsReadyPath       = "/downloads/uploads"
)

func MonitorDownloads(ctx context.Context, completed chan<- redisutils.Download) error {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
		return err
	}
	completedDownloads, err := c.GetCompletedDownloads(ctx)
	if err != nil {
		log.Printf("error getting completed downloads: %v", err)
		return err
	}

	rdb, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
	}
	defer rdb.Close()

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
			if err := redisutils.RegisterDownloadState(ctx, rdb, d); err != nil {
				log.Printf("error storing in redis: %v", err)
				continue
			}
			completed <- d
		}
	}
	return nil
}

func CompressDownload(ctx context.Context, download redisutils.Download) error {
	log.Printf("Compressing download: %s", download.Name)
	destination := filepath.Join(UploadsReadyPath, download.Name, download.Name)
	err := CompressAndSplitDownload(ctx, download.Path, destination)
	if err != nil {
		log.Printf("error compressing download: %v", err)
		return err
	}
	log.Printf("Compression completed: %v", download)
	return nil
}

// CompressionWorker to forward to upload channel
func CompressionWorker(ctx context.Context, compress <-chan redisutils.Download, upload chan<- redisutils.Download, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(upload)

	for {
		select {
		case <-ctx.Done():
			return
		case download, ok := <-compress:
			if !ok {
				return
			}

			// Compress file
			err := CompressDownload(ctx, download)
			if err != nil {
				log.Printf("error compressing %s: %v", download.Path, err)
				continue
			}

			// Forward to upload worker
			upload <- download
		}
	}
}

func UploadWorker(ctx context.Context, b *bot.Bot, upload <-chan redisutils.Download, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case download, ok := <-upload:
			if !ok {
				return
			}

			// Process upload here
			err := UploadDir(ctx, b, download)
			if err != nil {
				log.Printf("error uploading %s: %v", download.UploadPath, err)
			}

			log.Printf("Successfully uploaded: %s", download.UploadPath)

			// Remove uploaded files
			err = RemoveUploadedFiles(ctx, download)
			if err != nil {
				log.Printf("error removing uploaded files: %v", err)
			}

			// Remove torrent
			c, err := transmission.NewTransmissionClient(ctx)
			if err != nil {
				log.Printf("error creating transmission client: %v", err)
			}

			err = c.RemoveTorrents(ctx, []int64{download.ID})
			if err != nil {
				log.Printf("error removing torrent: %v", err)
			}
		}
	}
}

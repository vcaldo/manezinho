package fileutils

import (
	"context"
	"log"
	"path/filepath"

	"github.com/vcaldo/manezinho/bot/transmission"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
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
			Path:       filepath.Join(redisutils.ComplatedDownloadsPath, *download.Name),
			UploadPath: filepath.Join(redisutils.UploadsReadyPath, *download.Name),
			State:      redisutils.Downloaded,
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
			downloadChan <- d
		}
	}
	return nil
}

func ProcessDownload(ctx context.Context, download redisutils.Download) {
	log.Printf("Processing download %s", download.Name)

	rdb, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("error creating redis client: %v", err)
		return
	}
	defer rdb.Close()

	// Update a state to compressing in Redis
	download.State = redisutils.Compressing
	err = redisutils.RegisterDownloadState(ctx, rdb, download)
	if err != nil {
		log.Printf("error updating download state in redis: %v", err)
		return
	}

	// Compress download
	err = CompressDownload(ctx, download)
	if err != nil {
		log.Printf("error compressing and splitting download: %v", err)
		return
	}

	// Update a state to compressed in Redis
	download.State = redisutils.Compressed
	err = redisutils.RegisterDownloadState(ctx, rdb, download)
	if err != nil {
		log.Printf("error updating download state in redis: %v", err)
		return
	}

	log.Printf("Finished processing download %s", download.Name)
}

func CompressDownload(ctx context.Context, download redisutils.Download) error {
	log.Printf("Compressing download: %s", download.Name)
	destination := filepath.Join(redisutils.UploadsReadyPath, download.Name, download.Name)
	err := CompressAndSplitDownload(ctx, download.Path, destination)
	if err != nil {
		log.Printf("error compressing download: %v", err)
		return err
	}
	log.Printf("Compression completed: %v", download)
	return nil
}

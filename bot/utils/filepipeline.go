package utils

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/vcaldo/manezinho/bot/transmission"
)

type Download struct {
	ID         int64
	Name       string
	Path       string
	UploadPath string
}

func MonitorDownloads(ctx context.Context, completed chan<- Download) error {
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
	for _, download := range completedDownloads {
		completed <- Download{ID: *download.ID, Name: *download.Name, Path: fmt.Sprintf("/downloads/complete/%s", *download.Name), UploadPath: fmt.Sprintf("/downloads/upload/%s", *download.Name)}
	}
	return nil
}

func CompressDownload(ctx context.Context, download Download) error {
	log.Printf("Compressing download: %v\n", download)
	destination := fmt.Sprintf("/downloads/upload/%s/%s", download.Name, download.Name)
	err := CompressAndSplitDownload(ctx, download.Path, destination)
	if err != nil {
		log.Printf("error compressing download: %v", err)
		return err
	}
	log.Printf("Compression completed: %v\n", download)
	return nil
}

// CompressionWorker to forward to upload channel
func CompressionWorker(ctx context.Context, compress <-chan Download, upload chan<- Download, wg *sync.WaitGroup) {
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

func UploadWorker(ctx context.Context, b *bot.Bot, upload <-chan Download, wg *sync.WaitGroup) {
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
			err := UploaFile(ctx, b, download)
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
				return
			}

			err = c.RemoveTorrents(ctx, []int64{download.ID})
		}
	}
}

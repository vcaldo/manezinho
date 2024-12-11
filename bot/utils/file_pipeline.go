package utils

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/vcaldo/manezinho/bot/transmission"
	"github.com/vcaldo/manezinho/bot/utils/fileutils"
)

type Download struct {
	ID   int64
	Name string
	Path string
}

func MonitorDownloads(ctx context.Context, completed chan<- Download) {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		log.Printf("error creating transmission client: %v", err)
	}
	completedDownloads, err := c.GetCompletedDownloads(ctx)
	if err != nil {
		log.Printf("error getting completed downloads: %v", err)
	}
	for _, download := range completedDownloads {
		completed <- Download{ID: *download.ID, Name: *download.Name, Path: fmt.Sprintf("/downloads/complete/%s", *download.Name)}
	}
}

func CompressDownload(ctx context.Context, download Download) {
	log.Printf("Compressing download: %v\n", download)
	destination := fmt.Sprintf("/downloads/upload/%s/%s", download.Name, download.Name)
	err := fileutils.CompressDownload(ctx, download.Path, destination)
	if err != nil {
		log.Panicf("error compressing download: %v", err)
	}
	log.Printf("Compression completed: %v\n", download)
}

// CompressionWorker processes downloads one at a time
func CompressionWorker(ctx context.Context, compress <-chan Download, wg *sync.WaitGroup) {
	defer wg.Done()
	for download := range compress {
		CompressDownload(ctx, download)
	}
}

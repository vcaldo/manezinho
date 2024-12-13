package main

import (
	"context"
	"time"

	"github.com/vcaldo/manezinho/jonatas/fileutils"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

func main() {
	ctx := context.Background()
	downloadChan := make(chan redisutils.Download)

	// Goroutine to constantly check for completed downloads
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fileutils.GetCompletedDownloads(ctx, downloadChan)
			}
		}
	}()

	// Goroutine to process downloads one at a time
	go func() {
		for download := range downloadChan {
			fileutils.ProcessDownload(ctx, download)
		}
	}()

	// Keep the main function running
	select {}
}

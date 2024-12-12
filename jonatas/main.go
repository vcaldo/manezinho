package main

import (
	"context"
	"time"

	"github.com/vcaldo/manezinho/jonatas/redisutils"
	"github.com/vcaldo/manezinho/jonatas/utils"
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
				utils.GetCompletedDownloads(ctx, downloadChan)
			}
		}
	}()

	// Goroutine to process downloads one at a time
	go func() {
		for download := range downloadChan {
			utils.ProcessDownload(ctx, download)
		}
	}()

	// Keep the main function running
	select {}
}

package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/vcaldo/manezinho/metrics/utils"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// Check for completed downloads
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				utils.CheckCompletedDownloads(ctx)
			}
		}
	}()

	// Prevent the main function from exiting
	select {}
}

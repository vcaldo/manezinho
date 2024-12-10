package utils

import (
	"context"
	"log"

	"github.com/vcaldo/manezinho/bot/transmission"
	"github.com/vcaldo/manezinho/metrics/utils/redisutils"
)

func CheckCompletedDownloads(ctx context.Context) error {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		return nil
	}
	completedDownloads, err := c.GetCompletedDownloads(ctx)
	if err != nil {
		log.Printf("error getting completed downloads: %v", err)
	}

	r := redisutils.NewRedisClient(ctx)

	for _, download := range completedDownloads {
		log.Printf("download %s is completed. Id: %d", *download.Name, *download.ID)
		r.Lpush(ctx, "completed_downloads", *download.ID)
	}

	return nil
}

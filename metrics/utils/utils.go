package utils

import (
	"context"

	"github.com/vcaldo/manezinho/bot/transmission"
)

func CheckCompletedDownloads(ctx context.Context) error {
	c, err := transmission.NewTransmissionClient(ctx)
	if err != nil {
		return nil
	}
	c.GetCompletedDownloads(ctx)
	return nil

}

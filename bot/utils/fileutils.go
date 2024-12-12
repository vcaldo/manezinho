package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

func RemoveUploadedFiles(ctx context.Context, download redisutils.Download) error {
	if err := os.RemoveAll(download.UploadPath); err != nil {
		return fmt.Errorf("failed to remove uploaded files: %v", err)
	}
	return nil
}

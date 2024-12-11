package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func UploaFile(ctx context.Context, b *bot.Bot, download Download) error {
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}
	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}
	// Send the document
	files, err := os.ReadDir(download.UploadPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(download.UploadPath, file.Name())
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer file.Close()

			fileReader := &models.InputFileUpload{
				Filename: file.Name(),
				Data:     file,
			}

			_, err = b.SendDocument(context.Background(), &bot.SendDocumentParams{
				ChatID:   chatIdInt,
				Document: fileReader,
			})
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
		}
	}
	return nil
}

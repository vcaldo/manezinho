package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func UploadDir(ctx context.Context, b *bot.Bot, download Download) error {
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

			log.Printf("Uploading file: %s\n", file.Name())
			err = uploadFile(ctx, b, file, file.Name())
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
		}
	}
	return nil
}

func uploadFile(ctx context.Context, b *bot.Bot, file *os.File, fileName string) error {
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}

	// Send the document
	filePath := filepath.Join(UploadsPath, file.Name())
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	fileReader := &models.InputFileUpload{
		Filename: file.Name(),
		Data:     fileHandle,
	}

	_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   chatIdInt,
		Document: fileReader,
	})
	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}
	return nil
}

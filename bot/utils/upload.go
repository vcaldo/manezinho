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
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}

	files, err := os.ReadDir(download.UploadPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			file, err := os.Open(filepath.Join(download.UploadPath, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer file.Close()

			log.Printf("Uploading file: %s\n", file.Name())
			err = uploadFile(ctx, b, file, chatIdInt)
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
		}
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatIdInt,
		Text:   download.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

func uploadFile(ctx context.Context, b *bot.Bot, file *os.File, chatId int64) error {
	fileReader := &models.InputFileUpload{
		Filename: file.Name(),
		Data:     file,
	}

	_, err := b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   chatId,
		Document: fileReader,
	})
	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}
	return nil
}

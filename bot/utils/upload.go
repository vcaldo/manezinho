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

func UploaFile(ctx context.Context, b *bot.Bot, documentPath string) error {
	file, err := os.Open(documentPath)
	if err != nil {
		return fmt.Errorf("failed to open document: %v", err)
	}
	defer file.Close()

	// Get the base name of the document file
	fileName := filepath.Base(documentPath)

	// Create a file reader
	fileReader := &models.InputFileUpload{
		Filename: fileName,
		Data:     file,
	}
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}
	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}
	// Send the document
	_, err = b.SendDocument(context.Background(), &bot.SendDocumentParams{
		ChatID:   chatIdInt,
		Document: fileReader,
	})
	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}

	return nil
}

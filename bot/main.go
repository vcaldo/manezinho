package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/manezinho/bot/handlers"
	"github.com/vcaldo/manezinho/bot/utils"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	token, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		panic("BOT_TOKEN env var is required")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithServerURL(os.Getenv("LOCAL_TELEGRAM_BOT_API_URL")),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err.Error())
	}

	// Start the bot with a goroutine
	go func() {
		b.Start(ctx)
	}()

	uploadChan := make(chan redisutils.Download)

	// Goroutine to constantly check for completed downloads
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				utils.GetCompressedFiles(ctx, uploadChan)
			}
		}
	}()

	// Goroutine to process downloads one at a time
	go func() {
		chatId, ok := os.LookupEnv("CHAT_ID")
		if !ok {
			panic("CHAT_ID env var is required")
		}

		chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
		if err != nil {
			panic("CHAT_ID must be a valid int64")
		}

		for upload := range uploadChan {
			files, err := os.ReadDir(upload.UploadPath)
			if err != nil {
				log.Printf("failed to read directory: %v", err)
				return
			}

			for _, file := range files {
				if !file.IsDir() {
					file, err := os.Open(filepath.Join(upload.UploadPath, file.Name()))
					if err != nil {
						log.Printf("failed to open file: %v", err)
						return
					}
					defer file.Close()

					log.Printf("Uploading file: %s", file.Name())
					err = utils.UploadFile(ctx, b, file, chatIdInt)
					if err != nil {
						log.Printf("failed to upload file: %v", err)
						return
					}

					_, err = b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: chatIdInt,
						Text:   upload.Name,
					})
					if err != nil {
						log.Printf("failed to send message: %v", err)
						return
					}
				}
			}
		}
	}()

	select {}
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Check if the user is allowed
	if !handlers.IsUserAllowed(ctx, update.Message.From.ID) {
		b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo:  &models.InputFileString{Data: "https://ih1.redbubble.net/image.3655810608.7816/flat,750x,075,f-pad,750x1000,f8f8f8.jpg"},
			Caption: fmt.Sprintf(
				"⚠️ Access Restricted\n\n"+
					"This bot requires authorization for usage.\n"+
					"To request access, contact the administrator or the person who invited you, and provide your User ID\n"+
					"📋 User ID: %d\n\n"+
					"Thank you for your understanding.",
				update.Message.From.ID,
			),
		})
		return
	}
	// Switch case for handling different types of messages
	switch {
	// handle text message
	case update.Message != nil && update.Message.Text != "":
		handlers.HandleTextMessage(ctx, b, update)
		return
	// handle Documents
	case update.Message != nil && update.Message.Document != nil:
		handlers.HandleDocument(ctx, b, update)
		return
	}
}

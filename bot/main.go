package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/manezinho/bot/handlers"
	"github.com/vcaldo/manezinho/bot/redisutils"
	"github.com/vcaldo/manezinho/bot/utils"
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

	completed := make(chan redisutils.Download)
	compress := make(chan redisutils.Download)
	upload := make(chan redisutils.Download)

	var wg sync.WaitGroup

	// Start the compression worker and upload workers
	wg.Add(1)
	go utils.CompressionWorker(ctx, compress, upload, &wg)
	go utils.UploadWorker(ctx, b, upload, &wg)

	// Start monitoring downloads
	go utils.MonitorDownloads(ctx, completed)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				utils.MonitorDownloads(ctx, completed)
			}
		}
	}()

	// Signal compression worker with completed downloads
	for download := range completed {
		compress <- download
	}

	close(compress)
	wg.Wait()
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Check if the user is allowed
	if !utils.IsUserAllowed(ctx, update.Message.From.ID) {
		b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  update.Message.Chat.ID,
			Photo:   &models.InputFileString{Data: "https://ih1.redbubble.net/image.3655810608.7816/flat,750x,075,f-pad,750x1000,f8f8f8.jpg"},
			Caption: fmt.Sprintf("You are not allowed to use this bot. This incident will be reported.\nUser ID: %d", update.Message.From.ID),
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

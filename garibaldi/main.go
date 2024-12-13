package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

const (
	uploadPath = "/downloads/uploads"
)

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx := context.Background()
	// defer cancel()

	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		log.Fatal("CHAT_ID env var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		log.Fatal("CHAT_ID must be a valid int64")
	}

	token, ok := os.LookupEnv("BOT_UPLOAD_TOKEN")
	if !ok {
		log.Fatal("BOT_TOKEN env var is required")
	}

	opts := []bot.Option{
		bot.WithServerURL(os.Getenv("LOCAL_TELEGRAM_BOT_API_URL")),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Start the bot with a goroutine
	go func() {
		b.Start(ctx)
		log.Println("Bot started")
	}()

	rdb, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Fatalf("failed to create redis client: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done, exiting")
			return
		default:
			uploads, err := redisutils.GetDowloadState(ctx, rdb, redisutils.Compressed)
			if err != nil {
				log.Printf("failed to get uploads: %v", err)
				continue
			}

			for _, id := range uploads {
				name, err := redisutils.GetDownloadName(ctx, rdb, id)
				if err != nil {
					log.Printf("failed to get download name: %v", err)
					continue
				}

				files, err := os.ReadDir(filepath.Join(uploadPath, name))
				if err != nil {
					log.Printf("failed to open file: %v", err)
					continue
				}
				for _, file := range files {
					file, err := os.Open(filepath.Join(uploadPath, name, file.Name()))
					if err != nil {
						log.Printf("failed to open file: %v", err)
						continue
					}
					defer file.Close()

					fileReader := &models.InputFileUpload{
						Filename: file.Name(),
						Data:     file,
					}

					_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
						ChatID:   chatIdInt,
						Document: fileReader,
					})
					if err != nil {
						log.Printf("failed to send document: %v", err)
						return
					}
					log.Printf("Document sent: %s", file.Name())
					// b.Close(ctx)
				}
				_, err = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatIdInt,
					Text:   name,
				})
				if err != nil {
					log.Printf("failed to send message: %v", err)
					return
				}
				log.Printf("Message sent: %s", name)
			}
		}
	}
}

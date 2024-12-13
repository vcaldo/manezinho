package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

func GetCompressedFiles(ctx context.Context, uploadChan chan<- redisutils.Download) {
	rdb, err := redisutils.NewAuthenticatedRedisClient(ctx)
	if err != nil {
		log.Printf("failed to create redis client: %v", err)
		return
	}

	compressed, err := redisutils.GetDowloadState(ctx, rdb, redisutils.Compressed)
	if err != nil {
		log.Printf("failed to get downloads: %v", err)
		return
	}

	for _, id := range compressed {
		// Get Downlaod Name
		name, err := redisutils.GetDownloadName(ctx, rdb, id)
		if err != nil {
			log.Printf("failed to get download name: %v", err)
			continue
		}

		//get download path
		downloadPath, err := redisutils.GetDownloadPath(ctx, rdb, id)
		if err != nil {
			log.Printf("failed to get download path: %v", err)
			continue
		}

		// Get upload path
		uploadPath, err := redisutils.GetUploadPath(ctx, rdb, id)
		if err != nil {
			log.Printf("failed to get upload path: %v", err)
			continue
		}

		upload := redisutils.Download{
			ID:         id,
			Name:       name,
			UploadPath: uploadPath,
			Path:       downloadPath,
		}

		uploadChan <- upload
	}
}

func UploadDir(ctx context.Context, download redisutils.Download) error {
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

			log.Printf("Uploading file: %s", file.Name())
			err = uploadFile(ctx, file)
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func uploadFile(ctx context.Context, file *os.File) error {
	chatId, ok := os.LookupEnv("CHAT_ID")
	if !ok {
		panic("CHAT_ID env var is required")
	}

	chatIdInt, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic("CHAT_ID must be a valid int64")
	}

	token, ok := os.LookupEnv("BOT_UPLOAD_TOKEN")
	if !ok {
		panic("BOT_TOKEN env var is required")
	}

	opts := []bot.Option{
		bot.WithServerURL(os.Getenv("LOCAL_TELEGRAM_BOT_API_URL")),
	}

	uploadBot, err := bot.New(token, opts...)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		<-ctx.Done()
		uploadBot.Start(ctx)
	}()
	
	fileReader := &models.InputFileUpload{
		Filename: file.Name(),
		Data:     file,
	}

	_, err = uploadBot.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   chatIdInt,
		Document: fileReader,
	})

	if err != nil {
		return fmt.Errorf("failed to send document: %v", err)
	}
	return nil
}

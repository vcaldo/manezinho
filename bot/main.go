package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/vcaldo/manezinho/bot/handlers"
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

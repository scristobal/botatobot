package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"
)

// Send any text message to the bot after the bot has been started

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	token, ok := os.LookupEnv("BOT_TOKEN")

	if !ok {
		log.Fatal("BOT_TOKEN not found")
	}

	b := bot.New(token, opts...)

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}

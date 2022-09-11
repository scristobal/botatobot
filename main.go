package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/handlers"
	"scristobal/botatobot/queue"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {

	err := cfg.FromEnv()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Loading configuration...")

	opts := []bot.Option{
		bot.WithDefaultHandler(handlers.Update),
	}

	b := bot.New(cfg.BOT_TOKEN, opts...)

	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: []models.BotCommand{}})

	log.Println("Initializing job queue...")

	queue.Init(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				job := queue.Pop()
				handlers.Job(ctx, b, &job)
			}
		}
	}()

	log.Println("Starting bot...")

	b.Start(ctx)
}

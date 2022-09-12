package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/handlers"
	"scristobal/botatobot/worker"

	"github.com/go-telegram/bot"
)

func main() {
	log.Println("Loading configuration...")

	err := cfg.FromEnv()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Creating OS context...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Creating bot...")

	opts := []bot.Option{
		bot.WithDefaultHandler(handlers.Update),
	}

	b := bot.New(cfg.BOT_TOKEN, opts...)

	log.Println("Initializing job queue...")

	worker.Init(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				job := worker.Pop()
				handlers.Request(ctx, b, &job)
			}
		}
	}()

	log.Println("Starting bot...")

	b.Start(ctx)
}

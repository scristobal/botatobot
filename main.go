package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/handlers"
	"scristobal/botatobot/worker"
	"time"

	"github.com/go-telegram/bot"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	log.Println("Loading configuration...")

	err := cfg.FromEnv()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Creating OS context...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Initializing work queue...")

	queue := worker.Init[worker.Request](ctx)

	log.Println("Creating bot...")

	handlerUpdate := handlers.NewHandle(queue)

	opts := []bot.Option{
		bot.WithDefaultHandler(handlerUpdate),
	}

	b := bot.New(cfg.BOT_TOKEN, opts...)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				req := queue.Pop()
				handlers.Request(ctx, b, req)
			}
		}
	}()

	log.Println("Starting bot...")

	b.Start(ctx)
}

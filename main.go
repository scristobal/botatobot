package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/botatobot"
	"scristobal/botatobot/config"
	"time"

	"github.com/go-telegram/bot"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	log.Println("Loading configuration...")

	err := config.FromEnv()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Println("Creating OS context...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Initializing work queue...")

	queue := botatobot.NewQueue(ctx)

	log.Println("Starting bot...")

	handler := botatobot.NewHandler(queue)

	opts := []bot.Option{bot.WithDefaultHandler(handler)}

	b := bot.New(config.BOT_TOKEN, opts...)

	queue.RegisterBot(b)

	log.Println("listening to job queue...")
	go queue.Start()

	log.Println("listening to messages...")
	go b.Start(ctx)

	<-ctx.Done()
}

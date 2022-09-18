package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/config"
	botatobot "scristobal/botatobot/pkg"
	"time"

	"github.com/go-telegram/bot"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	if err := config.FromEnv(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	queue := botatobot.NewQueue()
	handler := botatobot.NewHandler(&queue)

	opts := []bot.Option{bot.WithDefaultHandler(handler)}
	b := bot.New(config.BOT_TOKEN, opts...)

	queue.RegisterBot(b)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go queue.Start(ctx)
	go b.Start(ctx)

	log.Println("Bot online, listening to messages...")

	<-ctx.Done()
}

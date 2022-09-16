package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"scristobal/botatobot/config"
	"scristobal/botatobot/handlers"
	"scristobal/botatobot/scheduler"
	"scristobal/botatobot/tasks"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

	requestFactory := func(m models.Message) ([]*scheduler.Runner[*tasks.Txt2img], error) {
		unwrappedTasks, err := tasks.FromString(m.Text)

		if err != nil {
			return nil, err
		}

		requests := make([]*scheduler.Runner[*tasks.Txt2img], len(unwrappedTasks))

		for i, t := range unwrappedTasks {
			requests[i] = &scheduler.Runner[*tasks.Txt2img]{Runner: t}
		}

		return requests, nil
	}

	queue := scheduler.NewQueue(ctx, requestFactory)

	log.Println("Starting bot...")

	handlerUpdate := handlers.NewHandler(queue)

	opts := []bot.Option{bot.WithDefaultHandler(handlerUpdate)}

	b := bot.New(config.BOT_TOKEN, opts...)

	queue.RegisterBot(b)

	log.Println("listening to job queue...")
	go queue.Start()

	log.Println("listening to messages...")
	go b.Start(ctx)

	<-ctx.Done()
}

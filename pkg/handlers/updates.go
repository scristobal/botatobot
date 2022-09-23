package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type queue interface {
	Push(item *models.Message) error
	Len() int
	IsWorking() bool
}

func getMessageOrUpdate(update *models.Update) (*models.Message, error) {

	if update == nil {
		return &models.Message{}, fmt.Errorf("empty update")
	}

	message := update.Message

	if message != nil {
		return message, nil
	}

	edited := update.EditedMessage

	if edited != nil {
		return edited, nil
	}

	return &models.Message{}, fmt.Errorf("no message found in update")
}

func Status(q queue) bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		if update == nil {
			log.Println("empty update")
			return
		}

		message := update.Message

		if message == nil {
			log.Println("empty message")
			return
		}

		isWorking := q.IsWorking()
		numJobs := q.Len()

		if !isWorking && numJobs == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   "I am doing nothing and there are no jobs in the queue 🤖",
			})
			return
		}

		if isWorking {

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   fmt.Sprintf("I am generating an image and the queue has %d more jobs", numJobs),
			})
			return
		}

		if numJobs > 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   fmt.Sprintf("I am doing nothing and the queue has %d more jobs. That's weird!! ", numJobs),
			})
			return
		}
	}
}

func Generate(q queue) bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		if update == nil {
			log.Println("empty update")
			return
		}

		message, err := getMessageOrUpdate(update)

		if err != nil {
			log.Printf("Failed to get message: %s", err)
			return
		}

		err = q.Push(message)

		log.Printf("Requested %s\n", message.Text)

		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           message.Chat.ID,
				Text:             fmt.Sprintf("Sorry, but your request was rejected 😬 %s", err),
				ReplyToMessageID: message.ID,
			})
			log.Printf("Requested %s but rejected by %s\n", message.Text, err)
			return
		}
	}

}

func Help() bot.HandlerFunc {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		if update == nil {
			log.Println("empty update")
			return
		}

		message := update.Message

		if message == nil {
			log.Println("empty message")
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
		})
	}
}

package botatobot

import (
	"context"
	"fmt"
	"log"

	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler = func(context.Context, *bot.Bot, *models.Update)

type queue interface {
	Push(item models.Message) error
	Len() int
	IsWorking() bool
}

func NewHandler(q queue) Handler {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in f", r)
			}
		}()

		message := update.Message

		if message == nil {
			log.Printf("Got an update without a message. Skipping. \n")
			return
		}

		text := update.Message.Text

		if strings.HasPrefix(text, string(Generate)) {
			generateHandler(ctx, b, *message, q)
		}

		if strings.HasPrefix(text, string(Help)) {
			helpHandler(ctx, b, message)
		}

		if strings.HasPrefix(text, string(Status)) {
			statusHandler(ctx, b, *message, q)
		}
	}
}

func statusHandler(ctx context.Context, b *bot.Bot, message models.Message, q queue) {
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

func generateHandler(ctx context.Context, b *bot.Bot, message models.Message, q queue) {

	err := q.Push(message)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but your request was rejected 😬\n\n %s", err),
			ReplyToMessageID: message.ID,
		})
		log.Printf("User %s requested %s but rejected", message.From.Username, err)
		return
	}

}

func helpHandler(ctx context.Context, b *bot.Bot, message *models.Message) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
	})
}

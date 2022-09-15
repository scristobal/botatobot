package controllers

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type statusQueue interface {
	Len() int
	IsWorking() bool
}

func Status(ctx context.Context, b *bot.Bot, message models.Message, q statusQueue) {
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

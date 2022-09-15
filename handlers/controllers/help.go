package controllers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Help(ctx context.Context, b *bot.Bot, message *models.Message) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
	})
}

package pkg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type command string

const (
	HelpCommand     command = "/help"
	GenerateCommand command = "/generate"
)

func GenerateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	if update == nil {
		log.Println("got an empty update, skipping")
		return
	}

	message := update.Message

	if message == nil {
		log.Printf("got no message from non-empty update, skipping")
		return
	}

	// prompt is message.Text after removing the command
	prompt := message.Text[(len(GenerateCommand) + 1):]
	prompt = strings.TrimSpace(prompt)

	// TODO: sanitize prompt

	log.Printf("User `%s` requested `%s`\n", message.Chat.Username, prompt)

	generator := ctx.Value(ImageGeneratorKey).(*imageGenerator)

	output, err := generator.GenerateImageFromPrompt(prompt)

	if err != nil {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but something went wrong 😭 %s", err),
			ReplyToMessageID: message.ID,
		})

		if err != nil {
			log.Printf("Error sending message: %s", err)
		}
		return
	}

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  message.Chat.ID,
		Caption: fmt.Sprint(prompt),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(output),
			Filename: filepath.Base(fmt.Sprintf("%s.png", uuid.New())),
		},
		DisableNotification: true,
	})

	if err != nil {
		log.Printf("Error sending photo: %s", err)
	}
}

func HelpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

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

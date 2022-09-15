package handlers

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Req interface {
	Id() uuid.UUID
	Msg() *models.Message
	Result() ([]byte, error)
}

func Request(ctx context.Context, b *bot.Bot, req Req) error {

	message := req.Msg()

	res, err := req.Result()

	if err != nil {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but something went wrong 😭 %s", err),
			ReplyToMessageID: message.ID,
		})

		return err
	}

	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  message.Chat.ID,
		Caption: fmt.Sprint(req),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(res),
			Filename: filepath.Base(fmt.Sprintf("%s.png", req.Id())),
		},
		DisableNotification: true,
	})

	return err
}

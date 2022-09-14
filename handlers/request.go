package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scristobal/botatobot/cfg"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Req interface {
	Id() uuid.UUID
	Msg() *models.Message
	Result() ([]byte, error)
}

func Request(ctx context.Context, b *bot.Bot, req Req) {

	message := req.Msg()

	res, err := req.Result()

	if err != nil {
		log.Printf("there was a problem running the task %s", err)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but something went wrong when running the model 😭 %s", err),
			ReplyToMessageID: message.ID,
		})

		return
	}

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  message.Chat.ID,
		Caption: fmt.Sprint(req),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(res),
			Filename: filepath.Base(fmt.Sprintf("%s.png", req.Id())),
		},
		DisableNotification: true,
	})

	imgFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.png", req.Id()))

	err = os.WriteFile(imgFilePath, res, 0644)

	if err != nil {
		log.Printf("can't write image to disc: %s\n", err)
	}

	content, err := json.Marshal(req)

	if err != nil {
		log.Printf("failed to serialize job parameters: %s\n", err)
	}

	jsonFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.json", req.Id()))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		log.Printf("error writing metadata: %s", err)
	}

}

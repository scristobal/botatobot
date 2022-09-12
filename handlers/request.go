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
	"scristobal/botatobot/worker"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Request(ctx context.Context, b *bot.Bot, req *worker.Request) {

	message := req.Msg

	if req.Error != nil {
		log.Printf("there was a problem running the task %s", req.Error)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but something went wrong when running the model 😭 %s", req.Error),
			ReplyToMessageID: message.ID,
		})

		return
	}

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  message.Chat.ID,
		Caption: fmt.Sprint(req),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(req.Result),
			Filename: filepath.Base(fmt.Sprintf("%s.png", req.Id)),
		},
		DisableNotification: true,
	})

	imgFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.png", req.Id))

	err := os.WriteFile(imgFilePath, req.Result, 0644)

	if err != nil {
		log.Printf("can't write image to disc: %s\n", err)
	}

	content, err := json.Marshal(req)

	if err != nil {
		log.Printf("failed to serialize job parameters: %s\n", err)
	}

	jsonFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.json", req.Id))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		log.Printf("error writing metadata: %s", err)
	}

}

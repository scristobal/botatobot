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

	requester := req.Requester

	for _, job := range req.Jobs {

		if job.Error != nil {
			log.Printf("Error job %s no file found\n", req.Id)

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           requester.ChatId,
				Text:             "Sorry, but something went wrong when running the model 😭",
				ReplyToMessageID: requester.MsgId,
			})

			return
		}

		b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  requester.ChatId,
			Caption: fmt.Sprint(job.String()),
			Photo: &models.InputFileUpload{
				Data:     bytes.NewReader(job.Result),
				Filename: filepath.Base(fmt.Sprintf("%s.png", req.Id)),
			},
			DisableNotification: true,
		})

		imgFilePath := filepath.Join(cfg.OUTPUT_PATH, fmt.Sprintf("%s.png", req.Id))

		err := os.WriteFile(imgFilePath, job.Result, 0644)

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

}

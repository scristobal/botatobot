package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/queue"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Job(ctx context.Context, b *bot.Bot, job *queue.Job) {

	imgPath := filepath.Join(cfg.OUTPUT_PATH, job.Id+".png")
	imgContent, err := os.ReadFile(imgPath)

	if err != nil {
		log.Printf("Error job %s no file found\n", job.Id)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           job.ChatId,
			Text:             "Sorry, but something went wrong when running the model 😭",
			ReplyToMessageID: job.MsgId,
		})

		return
	}

	log.Println("Success. Sending file from: ", job.Id)

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  job.ChatId,
		Caption: fmt.Sprint(job.Params),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(imgContent),
			Filename: filepath.Base(imgPath),
		},
		DisableNotification: true,
	})

}

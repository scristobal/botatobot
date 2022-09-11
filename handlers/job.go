package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/jobs"
	"scristobal/botatobot/queue"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Job(ctx context.Context, b *bot.Bot, job *queue.Job) {

	details := (*job).(jobs.Txt2img)

	imgPath := filepath.Join(cfg.OUTPUT_PATH, details.Id+".png")
	imgContent, err := os.ReadFile(imgPath)

	if err != nil {
		log.Printf("Error job %s no file found\n", details.Id)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           details.ChatId,
			Text:             "Sorry, but something went wrong when running the model 😭",
			ReplyToMessageID: details.MsgId,
		})

		return
	}

	log.Println("Success. Sending file from: ", details.Id)

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  details.ChatId,
		Caption: fmt.Sprint(details.Params),
		Photo: &models.InputFileUpload{
			Data:     bytes.NewReader(imgContent),
			Filename: filepath.Base(imgPath),
		},
		DisableNotification: true,
	})

}

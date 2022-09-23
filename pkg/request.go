package botatobot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"scristobal/botatobot/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Request struct {
	Task   Txt2img         `json:"task"`
	Id     uuid.UUID       `json:"id"`
	Msg    *models.Message `json:"msg"`
	Output []byte          `json:"-"`
	Err    error           `json:"error,omitempty"`
	Env    string          `json:"env"`
}

func (r *Request) GetIdentifier() uuid.UUID {
	return r.Id
}

func (r *Request) GetMessage() *models.Message {
	return r.Msg
}

func (r *Request) Launch() {
	r.Output, r.Err = r.Task.Execute(r.Env)
}

func (r *Request) Result() ([]byte, error) {
	return r.Output, r.Err
}

func (r *Request) String() string {
	return fmt.Sprintf("request %s with parameters: %s, running %s", r.Id, &r.Task, r.Env)
}

func (r *Request) SaveToDisk() error {

	if config.OUTPUT_PATH == "" {
		return fmt.Errorf("no output path defined, skipping")
	}

	err := os.MkdirAll(config.OUTPUT_PATH, 0755)

	if err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	if r.Err != nil {
		return fmt.Errorf("failed to get result: %s", err)
	}

	imgFilePath := filepath.Join(config.OUTPUT_PATH, fmt.Sprintf("%s.png", r.Id))

	err = os.WriteFile(imgFilePath, r.Output, 0644)

	if err != nil {
		return fmt.Errorf("can't write image to disc: %s", err)
	}

	content, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("failed to serialize job parameters: %s", err)
	}

	jsonFilePath := filepath.Join(config.OUTPUT_PATH, fmt.Sprintf("%s.json", r.Id))

	err = os.WriteFile(jsonFilePath, content, 0644)

	if err != nil {
		return fmt.Errorf("error writing metadata: %s", err)
	}
	return nil
}

func (r *Request) GetDescription() string {
	return r.Task.String()
}

type outcome interface {
	GetMessage() *models.Message
	GetIdentifier() uuid.UUID
	GetDescription() string
	Result() ([]byte, error)
}

func SendOutcome(ctx context.Context, b *bot.Bot) func(req outcome) error {
	return func(req outcome) error {

		message := req.GetMessage()

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
			Caption: fmt.Sprint(req.GetDescription()),
			Photo: &models.InputFileUpload{
				Data:     bytes.NewReader(res),
				Filename: filepath.Base(fmt.Sprintf("%s.png", req.GetIdentifier())),
			},
			DisableNotification: true,
		})

		return err
	}
}

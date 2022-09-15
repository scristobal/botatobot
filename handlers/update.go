package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"scristobal/botatobot/config"
	"scristobal/botatobot/handlers/controllers"
	"scristobal/botatobot/queue"
	"scristobal/botatobot/requests"

	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type Command string

// commands
const (
	Help     Command = "/help"
	Generate Command = "/generate"
	Status   Command = "/status"
)

func (c Command) String() string {
	switch c {
	case Help:
		return "Get Help, usage /help"
	case Generate:
		return "Generate a image from a prompt, usage /generate <prompt>"
	case Status:
		return "Check bot status, usage /status"
	}

	return "Unknown command"
}

type job interface {
	Run()
	Describe() string
	Result() ([]byte, error)
}

type Factory[T job] func(models.Message) ([]requests.Request[T], error)

func NewHandle[T job](q queue.Queue[requests.Request[T]], factory Factory[T]) func(context.Context, *bot.Bot, *models.Update) {

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {

		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in f", r)
			}
		}()

		message := update.Message

		if message == nil {
			log.Printf("Got an update without a message. Skipping. \n")
			return
		}

		m := update.Message.Text

		id := uuid.New()

		if strings.HasPrefix(m, string(Generate)) {

			requests, err := factory(*message)

			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:           message.Chat.ID,
					Text:             fmt.Sprintf("Sorry, but your request is somehow invalid 😬\n\n %s", err),
					ReplyToMessageID: message.ID,
				})
				log.Printf("User %s requested %s but rejected", message.From.Username, err)
				return
			}

			log.Printf("User %s requested %s accepted\n", message.From.Username, id)

			if q.Len() >= config.MAX_JOBS {
				b.SendMessage(ctx,
					&bot.SendMessageParams{
						ChatID:           message.Chat.ID,
						Text:             "Sorry, but the job queue reached its maximum, try again later 🙄",
						ReplyToMessageID: message.ID,
					})

				log.Println("User", message.From.Username, "request rejected, queue full")
				return
			}

			log.Printf("User %s request accepted, job id %s", message.From.Username, id)

			for _, req := range requests {
				req := req
				q.Push(req)
			}
		}

		if strings.HasPrefix(m, string(Help)) {
			controllers.Help(ctx, b, message)
		}

		if strings.HasPrefix(m, string(Status)) {
			job := q.Current()

			numJobs := q.Len()

			if job == nil && numJobs == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: message.Chat.ID,
					Text:   "I am doing nothing and there are no jobs in the queue 🤖",
				})
				return
			}

			if job != nil {

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

		if strings.HasPrefix(m, "/video-test") {

			prompt := strings.ReplaceAll(m, "/video-test", "")

			host := "http://127.0.0.1:5000/predictions"

			res, err := http.Post(host, "application/json", strings.NewReader(fmt.Sprintf(
				`{"input": {
				"max_frames": 300,
				"animation_prompts": "0: %s",
				"angle": "0:(0)",
				"zoom": "0: (1)",
				"translation_x": "0: (0)",
				"translation_y": "0: (0)",
				"color_coherence": "Match Frame 0 LAB",
				"sampler": "plms",
				"fps": 10,
				"seed": 242351
			}}`,
				prompt,
			)))

			if err != nil {
				log.Println(err)
				return
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)

			if err != nil {
				log.Println(err)
				return
			}

			type modelResponse struct {
				Status string `json:"status"`
				Output string `json:"output"`
			}

			response := modelResponse{}

			err = json.Unmarshal(body, &response)

			if err != nil {
				log.Println(err)
				return
			}

			data := strings.SplitAfter(response.Output, ",")[1]

			decoded, err := base64.StdEncoding.DecodeString(data)

			if err != nil {
				log.Println("Error decoding base64: ", err)

				return
			}

			fileName := fmt.Sprintf("%s.mp4", strings.ReplaceAll(prompt, " ", "_"))

			os.WriteFile(fileName, decoded, 0644)

			msg, err := b.SendVideo(
				ctx,
				&bot.SendVideoParams{
					ChatID:  message.Chat.ID,
					Caption: "Test video",
					Video:   &models.InputFileUpload{Filename: "sample.mp4", Data: bytes.NewReader(decoded)},
				})

			if err != nil {
				log.Println("Error sending video: ", err, msg)
				return
			}

		}

	}
}

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
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/worker"
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

func Update(ctx context.Context, b *bot.Bot, update *models.Update) {

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

	requester := worker.Requester{
		Msg:    update.Message.Text,
		MsgId:  update.Message.ID,
		ChatId: update.Message.Chat.ID,
		User:   update.Message.From.Username,
		UserId: update.Message.From.ID,
	}

	id := uuid.New()

	if strings.HasPrefix(m, string(Generate)) {

		req, err := worker.New(requester)

		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           requester.ChatId,
				Text:             fmt.Sprintf("Sorry, but your request is somehow invalid 😬\n\n %s", err),
				ReplyToMessageID: requester.MsgId,
			})
			log.Printf("User %s requested %s but rejected", requester.User, err)
			return
		}

		log.Printf("User %s requested %s accepted\n", requester.User, id)

		if worker.Len() >= cfg.MAX_JOBS {
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID:           requester.ChatId,
					Text:             "Sorry, but the job queue reached its maximum, try again later 🙄",
					ReplyToMessageID: requester.MsgId,
				})

			log.Println("User", requester.User, "request rejected, queue full")
			return
		}

		log.Printf("User %s request accepted, job id %s", requester.User, id)

		for _, job := range req.Jobs {
			worker.Push(&job)
		}
	}

	if strings.HasPrefix(m, string(Help)) {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: requester.ChatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
		})
	}

	if strings.HasPrefix(m, string(Status)) {

		job := worker.Current()

		numJobs := worker.Len()

		if job == nil && numJobs == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: requester.ChatId,
				Text:   "I am doing nothing and there are no jobs in the queue 🤖",
			})
			return
		}

		if job != nil {

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: requester.ChatId,
				Text:   fmt.Sprintf("I am generating an image and the queue has %d more jobs", numJobs),
			})
			return
		}

		if numJobs > 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: requester.ChatId,
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
				ChatID:  requester.ChatId,
				Caption: "Test video",
				Video:   &models.InputFileUpload{Filename: "sample.mp4", Data: bytes.NewReader(decoded)},
			})

		if err != nil {
			log.Println("Error sending video: ", err, msg)
			return
		}

	}

}

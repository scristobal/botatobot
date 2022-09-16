package botatobot

import (
	"context"
	"fmt"
	"log"

	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

type queue interface {
	Push(item models.Message) error
	Len() int
	IsWorking() bool
}

type Handler = func(context.Context, *bot.Bot, *models.Update)

func NewHandler(q queue) Handler {

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

		if strings.HasPrefix(m, string(Generate)) {
			generateHandler(ctx, b, *message, q)
		}

		if strings.HasPrefix(m, string(Help)) {
			helpHandler(ctx, b, message)
		}

		if strings.HasPrefix(m, string(Status)) {
			statusHandler(ctx, b, *message, q)
		}
		/*
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
		*/
	}
}

func statusHandler(ctx context.Context, b *bot.Bot, message models.Message, q queue) {
	isWorking := q.IsWorking()

	numJobs := q.Len()

	if !isWorking && numJobs == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "I am doing nothing and there are no jobs in the queue 🤖",
		})
		return
	}

	if isWorking {

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

func generateHandler(ctx context.Context, b *bot.Bot, message models.Message, q queue) {

	err := q.Push(message)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           message.Chat.ID,
			Text:             fmt.Sprintf("Sorry, but your request was rejected 😬\n\n %s", err),
			ReplyToMessageID: message.ID,
		})
		log.Printf("User %s requested %s but rejected", message.From.Username, err)
		return
	}

}
func helpHandler(ctx context.Context, b *bot.Bot, message *models.Message) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
	})
}

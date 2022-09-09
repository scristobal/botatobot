package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"scristobal/botatobot/cfg"
	"scristobal/botatobot/cmd"
	"scristobal/botatobot/worker"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/google/uuid"
)

func main() {

	err := cfg.FromEnv()

	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Loading configuration...")

	opts := []bot.Option{
		bot.WithDefaultHandler(handleUpdate),
	}

	b := bot.New(cfg.BOT_TOKEN, opts...)

	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: []models.BotCommand{}})

	log.Println("Initializing job queue...")

	worker.Init(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				job := worker.Pop()
				handleJob(ctx, b, &job)
			}
		}
	}()

	log.Println("Starting bot...")

	b.Start(ctx)
}

func handleJob(ctx context.Context, b *bot.Bot, job *worker.Job) {

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

func handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {

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

	messageText := update.Message.Text
	messageId := update.Message.ID

	chat := update.Message.Chat

	chatId := chat.ID

	user := update.Message.From.Username
	userId := update.Message.From.ID

	id := uuid.New()

	if strings.HasPrefix(messageText, string(cmd.Generate)) {

		params, hasParams, err := cmd.GetParams(messageText)

		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           chatId,
				Text:             fmt.Sprintf("Sorry, but your prompt is somehow invalid 😬\n\n %s", err),
				ReplyToMessageID: messageId,
			})
			log.Printf("Invalid prompt from %s: %s", user, err)
			return
		}

		log.Printf("User %s requested %s \n", user, params.Prompt)

		if worker.Len() >= cfg.MAX_JOBS {
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID:           chatId,
					Text:             "Sorry, but the job queue reached its maximum, try again later 🙄",
					ReplyToMessageID: messageId,
				})

			log.Println("User", user, "request rejected, queue full")
			return
		}

		log.Printf("User %s request accepted, job id %s", user, id)

		if !hasParams {

			for i := 0; i < 5; i++ {
				params.Seed = rand.Intn(1000000)
				worker.Push(worker.Job{ChatId: chatId, User: user, UserId: userId, MsgId: messageId, Id: id.String(), Params: params})
			}
			return
		}

		worker.Push(worker.Job{ChatId: chatId, User: user, UserId: userId, MsgId: messageId, Id: id.String(), Params: params})
	}

	if strings.HasPrefix(messageText, string(cmd.Help)) {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nBy default I will generate 5 images, but you can modify the seed, guidance and steps like so\n\n /generate a cat in space &seed_1234 &steps_50 &guidance_7.5\n\nCheck my status with /status\n\nHave fun!",
		})
	}

	if strings.HasPrefix(messageText, string(cmd.Status)) {

		job := worker.Current()

		numJobs := worker.Len()

		if job == nil && numJobs == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   "I am doing nothing and there are no jobs in the queue 🤖",
			})
			return
		}

		if job != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   fmt.Sprintf("I am working on \"%s\" for %s and the queue has %d more jobs", job.Params, job.User, numJobs),
			})
			return
		}

		if numJobs > 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   fmt.Sprintf("I am doing nothing and the queue has %d more jobs. That's weird!! ", numJobs),
			})
			return
		}
	}

	if strings.HasPrefix(messageText, "/video-test") {

		prompt := strings.ReplaceAll(messageText, "/video-test", "")

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
				ChatID:  chatId,
				Caption: "Test video",
				Video:   &models.InputFileUpload{Filename: "sample.mp4", Data: bytes.NewReader(decoded)},
			})

		if err != nil {
			log.Println("Error sending video: ", err, msg)
			return
		}

	}

}

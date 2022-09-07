package main

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
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"

	"github.com/google/uuid"
	"golang.org/x/exp/utf8string"
)

var (
	BOT_TOKEN    string
	BOT_USERNAME string
	MODEL_PATH   string
	OUTPUT_PATH  string
)

type Command int8

// commands
const (
	Help Command = iota
	Generate
	Status
)

func (c Command) String() string {
	switch c {
	case Help:
		return "/help"
	case Generate:
		return "/generate"

	case Status:
		return "/status"
	}
	return "unknown"
}

const MAX_JOBS = 10

type job struct {
	ChatId int
	Prompt string
	User   string
	UserId int
	MsgId  int
	Id     string
}

var jobQueue = make(chan job, MAX_JOBS)

type jobResult struct {
	Job     job
	Err     error
	Outputs []string
}

var jobResults = make(chan jobResult, MAX_JOBS)

var currentJob struct {
	job *job
	mut sync.RWMutex
}

func configure() error {
	err := godotenv.Load()
	var ok bool

	if err != nil {
		log.Println("Failed to load .env file, fallback on env vars")
	}

	BOT_TOKEN, ok = os.LookupEnv("BOT_TOKEN")

	if !ok {
		return fmt.Errorf("BOT_TOKEN not found")
	}

	BOT_USERNAME, ok = os.LookupEnv("BOT_USERNAME")

	if !ok {
		return fmt.Errorf("BOT_USERNAME not found")
	}

	MODEL_PATH, ok = os.LookupEnv("MODEL_PATH")

	if !ok {
		return fmt.Errorf("MODEL_PATH not found")
	}

	OUTPUT_PATH, ok = os.LookupEnv("OUTPUT_PATH")

	if !ok {
		return fmt.Errorf("OUTPUT_PATH not found")
	}

	return nil

}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Loading configuration...")

	err := configure()

	if err != nil {
		log.Fatal(err)
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b := bot.New(BOT_TOKEN, opts...)

	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{
				Command:     Generate.String(),
				Description: "Generate a text from a prompt",
			},
			{
				Command:     Status.String(),
				Description: "Check the status of the current job, if any, and the queue length",
			},
			{
				Command:     Help.String(),
				Description: "Get help",
			},
		}})

	log.Println("Initializing job queue...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-jobQueue:
				jobResults <- processJobs(job)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-jobResults:
				resolveJob(ctx, b, result)
			}
		}
	}()

	log.Println("Starting bot...")

	b.Start(ctx)

}

func validate(prompt string) bool {

	ok := utf8string.NewString(prompt).IsASCII()

	if !ok {
		return false
	}

	re := regexp.MustCompile(`^[\w\d\s-:.]*$`)

	return re.MatchString(prompt) && len(prompt) > 0
}

func clean(msg string) string {
	msg = strings.ReplaceAll(msg, BOT_USERNAME, "")

	msg = strings.ReplaceAll(msg, Generate.String(), "")

	msg = strings.ReplaceAll(msg, "\"", "")

	msg = strings.ReplaceAll(msg, ",", " ")

	msg = strings.ReplaceAll(msg, "_", " ")

	msg = strings.ReplaceAll(msg, "!", " ")

	msg = strings.ReplaceAll(msg, "?", " ")

	msg = strings.TrimSpace(msg)

	// removes consecutive spaces
	reg := regexp.MustCompile(`\s+`)
	msg = reg.ReplaceAllString(msg, " ")

	return msg
}

func getPrompt(msg string) (string, error) {

	prompt := clean(msg)

	ok := validate(prompt)

	if !ok {
		return "", fmt.Errorf("invalid characters in prompt")
	}

	return prompt, nil
}

func Mention(name string, id int) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", name, id)
}

func processJobs(job job) jobResult {

	currentJob.mut.Lock()
	currentJob.job = &job
	currentJob.mut.Unlock()

	outputFolder := fmt.Sprintf("%s/%s", OUTPUT_PATH, job.Id)

	args := []string{"-i", "run_sd.sh", job.Prompt, outputFolder}

	cmd := exec.Command("zsh", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	cmd.Env = append(os.Environ(), fmt.Sprintf("MODEL_PATH=%s", MODEL_PATH))

	err := cmd.Run()

	if err != nil {
		log.Printf("Error running script: %v", err)
		return jobResult{Job: job, Err: err}

	}

	var outputPaths []string

	filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			outputPaths = append(outputPaths, path)
		}
		return nil
	})

	currentJob.mut.Lock()
	currentJob.job = nil
	currentJob.mut.Unlock()

	return jobResult{Job: job, Outputs: outputPaths}
}

func resolveJob(ctx context.Context, b *bot.Bot, result jobResult) {

	content, err := json.Marshal(result)

	if err != nil {
		log.Printf("Error marshalling job result: %v", err)
	}

	outputFolder := fmt.Sprintf("%s/%s", OUTPUT_PATH, result.Job.Id)

	err = os.WriteFile(fmt.Sprintf("%s/meta.json", outputFolder), content, 0644)

	if err != nil {
		log.Printf("Error writing meta.json of job %s: %v", result.Job.Id, err)
	}

	if result.Err != nil {
		log.Printf("Failed to run %s, error: %v\n", result.Job.Id, result.Err)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:           result.Job.ChatId,
			Text:             "Sorry, but something went wrong when running the model 😭",
			ReplyToMessageID: result.Job.MsgId,
		})

		return
	}

	log.Println("Success. Sending files from: ", result.Job.Id)

	var media []models.InputMedia

	for _, output := range result.Outputs {

		fileContent, _ := os.ReadFile(output)

		media = append(media, &models.InputMediaPhoto{
			Media:           fmt.Sprintf("attach://%s", output),
			MediaAttachment: bytes.NewReader(fileContent),
			Caption:         result.Job.Prompt,
		})
	}

	b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
		ChatID:              result.Job.ChatId,
		Media:               media,
		DisableNotification: true,
		ReplyToMessageID:    result.Job.MsgId,
	})

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

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

	if strings.HasPrefix(messageText, Generate.String()) {

		prompt, err := getPrompt(messageText)

		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           chatId,
				Text:             "Sorry, but your prompt is somehow invalid 😬",
				ReplyToMessageID: messageId,
			})
			log.Printf("Invalid prompt from %s: %s", user, err)
			return
		}

		log.Printf("User %s requested %s \n", user, prompt)

		if len(jobQueue) >= MAX_JOBS {
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID:           chatId,
					Text:             "Sorry, but the job queue reached its maximum, try again later 🙄",
					ReplyToMessageID: messageId,
				})

			log.Println("User", user, "request rejected, queue full")
			return
		}

		/*
			b.SendMessage(ctx,
			&bot.SendMessageParams{
				ChatID:              chatId,
				Text:                "Your request is being processed 🤖",
				ReplyToMessageID:    messageId,
				DisableNotification: true,
			})
		*/

		log.Printf("User %s request accepted, job id %s", user, id)

		jobQueue <- job{ChatId: chatId, Prompt: prompt, User: user, UserId: userId, MsgId: messageId, Id: id.String()}

		return
	}

	if strings.HasPrefix(messageText, Help.String()) {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Use the /generate command follow by a prompt, like this: \n\n   /generate a cat in space \n\nand I'll generate a few images for you!. It can take a while, you can check my status with /status",
		})
	}

	if strings.HasPrefix(messageText, Status.String()) {

		currentJob.mut.RLock()
		defer currentJob.mut.RUnlock()

		numJobs := len(jobQueue)

		if currentJob.job == nil && numJobs == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   "I am doing nothing and there are no jobs in the queue 🤖",
			})
			return
		}

		if currentJob.job != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:           chatId,
				Text:             fmt.Sprintf("I am working on this message and the queue has %d more jobs", len(jobQueue)),
				ReplyToMessageID: currentJob.job.MsgId,
			})
			return
		}

		if numJobs > 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   fmt.Sprintf("I am doing nothing and the queue has %d more jobs. That's weird!! ", len(jobQueue)),
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
				"translation_x": "0: (5)",
				"translation_y": "0: (0)",
				"color_coherence": "Match Frame 0 LAB",
				"sampler": "plms",
				"fps": 15,
				"seed": 142351
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

	if strings.HasPrefix(messageText, "/photo-test") {

		prompt := strings.ReplaceAll(messageText, "/video-test", "")

		host := "http://127.0.0.1:5001/predictions"

		type modelResponse struct {
			Status string   `json:"status"`
			Output []string `json:"output"`
		}

		res, err := http.Post(host, "application/json", strings.NewReader(fmt.Sprintf(
			`{"input": {
				"prompt": "%s",
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

		response := modelResponse{}

		json.Unmarshal(body, &response)

		// it returns a (base64) data URL
		// get only first image
		data := strings.SplitAfter(response.Output[0], ",")[1]

		decoded, err := base64.StdEncoding.DecodeString(data)

		if err != nil {
			log.Println("Error decoding base64: ", err)
			return
		}

		fileName := fmt.Sprintf("%s.mp4", strings.ReplaceAll(prompt, " ", "_"))

		os.WriteFile(fileName, decoded, 0644)

		msg, err := b.SendPhoto(
			ctx,
			&bot.SendPhotoParams{
				ChatID:  chatId,
				Caption: "Test photo",
				Photo:   &models.InputFileUpload{Filename: "sample.png", Data: bytes.NewReader(decoded)},
			})

		if err != nil {
			log.Println("Error sending photo: ", err, msg)
			return
		}

	}

}

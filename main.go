package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"

	"github.com/google/uuid"
	"golang.org/x/exp/utf8string"
)

const MAGIC_WORDS = "@BotatoideBot"

var (
	BOT_TOKEN   string
	SCRIPT_PATH string
	OUTPUT_PATH string
)

const MAX_JOBS = 10

type job struct {
	chatId int
	prompt string
	user   string
	userId int
	id     string
}

var jobQueue = make(chan job, MAX_JOBS)

type jobResult struct {
	job    job
	err    error
	output string
}

var jobResults = make(chan jobResult, MAX_JOBS)

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

	SCRIPT_PATH, ok = os.LookupEnv("MODEL_PATH")

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

	err := configure()

	if err != nil {
		log.Fatal(err)
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b := bot.New(BOT_TOKEN, opts...)

	log.Println("Config loaded. Bot is running")

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

	b.Start(ctx)
}

func validate(prompt string) bool {

	ok := utf8string.NewString(prompt).IsASCII()

	if !ok {
		return false
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9, ]*$`)

	return re.MatchString(prompt)
}

func getPrompt(msg string) (string, error) {

	prompt := strings.Replace(msg, MAGIC_WORDS, "", -1)

	ok := validate(prompt)

	if !ok {
		return "", fmt.Errorf("invalid characters in prompt")
	}

	return prompt, nil
}

func mention(name string, id int) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", name, id)
}

func processJobs(job job) jobResult {

	outputFolder := fmt.Sprintf("%s/%s", OUTPUT_PATH, job.id)

	args := []string{"-i", SCRIPT_PATH, job.prompt, outputFolder}

	cmd := exec.Command("zsh", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Printf("Error running script: %v", err)
		return jobResult{job: job, err: err}

	}

	var outputPath string

	filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			outputPath = path
		}
		return nil
	})

	return jobResult{job: job, output: outputPath}
}

func resolveJob(ctx context.Context, b *bot.Bot, result jobResult) {

	if result.err != nil {
		log.Println("Failed to run the model")

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: result.job.chatId,
			Text:   fmt.Sprintf("Sorry %s, something went wrong when running the model 😭", mention(result.job.user, result.job.userId)),
		})

		return
	}

	log.Println("Success. Sending file: ", result.output)

	fileContent, _ := os.ReadFile(result.output)

	params := &bot.SendPhotoParams{
		ChatID: result.job.chatId,
		Photo:  &models.InputFileUpload{Filename: "image.png", Data: bytes.NewReader(fileContent)},
	}

	b.SendPhoto(ctx, params)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    result.job.chatId,
		Text:      fmt.Sprintf("%s by %s", result.job.prompt, mention(result.job.user, result.job.userId)),
		ParseMode: "Markdown",
	})

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	chatId := update.Message.Chat.ID
	message := update.Message.Text
	user := update.Message.From.Username
	userId := update.Message.From.ID
	id := uuid.New()

	if strings.HasPrefix(message, MAGIC_WORDS) {

		prompt, ok := getPrompt(message)

		if ok != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    chatId,
				Text:      fmt.Sprintf("Sorry %s your prompt is somehow invalid 😬", mention(user, userId)),
				ParseMode: "Markdown",
			})
			log.Println("Invalid prompt from", user)
			return
		}

		log.Printf("User %s requested %s \n", user, prompt)

		if len(jobQueue) >= MAX_JOBS {
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID:    chatId,
					Text:      fmt.Sprintf("Sorry %s, the job queue reached its maximum, try again later 🙄", mention(user, userId)),
					ParseMode: "Markdown",
				})

			log.Println("User", user, "request rejected, queue full")
			return
		}

		b.SendMessage(ctx,
			&bot.SendMessageParams{
				ChatID:    chatId,
				Text:      fmt.Sprintf("%s, your request is being processed 🤖", mention(user, userId)),
				ParseMode: "Markdown",
			})

		log.Println("User", user, "request accepted")

		jobQueue <- job{chatId: chatId, prompt: prompt, user: user, userId: userId, id: id.String()}

		return
	}

	if strings.HasPrefix(message, MAGIC_WORDS+" /help") {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Just mention me follow by a prompt, like this: \n\n @BotatoideBot a cat in space \n\n and I'll generate an image for you!",
		})
	}

	if strings.HasPrefix(message, MAGIC_WORDS+" /queue") {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   fmt.Sprintf("The queue has %d jobs", len(jobQueue)),
		})
	}

}

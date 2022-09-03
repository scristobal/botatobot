package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"

	"golang.org/x/exp/utf8string"
)

const MAGIC_WORDS = "@BotatoideBot"

var (
	BOT_TOKEN   string
	SCRIPT_PATH string
)

const MAX_JOBS = 10

type job struct {
	chatID int
	prompt string
	User   string
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

func processJobs(job job) jobResult {

	args := []string{"-i", SCRIPT_PATH, job.prompt}

	cmd := exec.Command("zsh", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Printf("Error running script: %v", err)
		return jobResult{job: job, err: err}

	}

	folderName := strings.Replace(job.prompt, " ", "_", -1)

	if len(folderName) > 126 {
		folderName = folderName[:126]
	}

	outputPath := strings.ReplaceAll(SCRIPT_PATH, "/run_sd.sh", "") + "/outputs/txt2img-samples/" + folderName + "/seed_27_00000.png"

	return jobResult{job: job, output: outputPath}
}

func resolveJob(ctx context.Context, b *bot.Bot, result jobResult) {

	if result.err != nil {
		log.Println("Failed to run the model")

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: result.job.chatID,
			Text:   fmt.Sprintf("Something went wrong when running the prompt %s for %s 😭", result.job.prompt, result.job.User),
		})

		return
	}

	log.Println("Success. Sending file: ", result.output)

	fileContent, _ := os.ReadFile(result.output)

	params := &bot.SendPhotoParams{
		ChatID:  result.job.chatID,
		Photo:   &models.InputFileUpload{Filename: "image.png", Data: bytes.NewReader(fileContent)},
		Caption: fmt.Sprintf("%s by %s", result.job.prompt, result.job.User),
	}

	b.SendPhoto(ctx, params)

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

	if strings.HasPrefix(message, MAGIC_WORDS) {

		prompt, ok := getPrompt(message)

		if ok != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   "Sorry your prompt is somehow invalid 😬"})
			log.Println("Invalid prompt from", user)
			return
		}

		log.Printf("User %s requested %s \n", user, prompt)

		if len(jobQueue) >= MAX_JOBS {
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID: chatId,
					Text:   "The job queue reached its maximum, try again later 🙄",
				})

			log.Println("User", user, "request rejected, queue full")
			return
		}

		b.SendMessage(ctx,
			&bot.SendMessageParams{
				ChatID: chatId,
				Text:   fmt.Sprintf("%s, your request is being processed 🤖", user),
			})

		log.Println("User", user, "request accepted")

		jobQueue <- job{chatID: chatId, prompt: prompt, User: user}

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

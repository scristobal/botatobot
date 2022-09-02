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
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"

	"golang.org/x/exp/utf8string"
)

const MAGIC_WORDS = "@BotatoideBot show me "

type sd_path string

type config struct {
	token string
	path  string
}

func (c *config) load() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file, fallback on env vars")
	}

	token, ok := os.LookupEnv("BOT_TOKEN")

	if !ok {
		log.Fatal("BOT_TOKEN not found")
	}

	path, ok := os.LookupEnv("SD_PATH")

	if !ok {
		log.Fatal("SD_PATH not found")
	}

	c.token = token
	c.path = path
}

type workLockerKey string

type workLocker struct {
	working bool
	mut     sync.RWMutex
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	c := config{}

	c.load()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b := bot.New(c.token, opts...)

	ctx = context.WithValue(ctx, sd_path("sd_path"), c.path)

	lock := &workLocker{working: false}

	ctx = context.WithValue(ctx, workLockerKey("workLocker"), lock)

	fmt.Println("Config loaded. Bot is running")

	b.Start(ctx)
}

func generate(ctx context.Context, prompt string) (string, bool, error) {

	lock := ctx.Value(workLockerKey("workLocker")).(*workLocker)

	lock.mut.RLock()
	if lock.working {
		lock.mut.RUnlock()
		return "", true, fmt.Errorf("already working")
	} else {
		lock.mut.RUnlock()
	}

	lock.mut.Lock()
	lock.working = true
	lock.mut.Unlock()

	modelPath, ok := ctx.Value(sd_path("sd_path")).(string)

	if !ok {
		log.Fatal("SD_PATH not found")
	}

	outputPath := modelPath + "/outputs/txt2img-samples/" + strings.Replace(prompt, " ", "_", -1) + "/seed_27_00000.png"

	args := []string{"-i", "run_sd.sh", modelPath, prompt}

	cmd := exec.Command("zsh", args...)

	err := cmd.Run()

	lock.mut.Lock()
	lock.working = false
	lock.mut.Unlock()

	return outputPath, false, err

}

func validate(prompt string) bool {

	ok := utf8string.NewString(prompt).IsASCII()

	if !ok {
		return false
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9_ ]*$`)

	return re.MatchString(prompt)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	chatId := update.Message.Chat.ID
	message := update.Message.Text
	user := update.Message.From.Username

	if strings.HasPrefix(message, MAGIC_WORDS) {

		messageBody := strings.Replace(message, MAGIC_WORDS, "", -1)

		prompt := strings.Replace(messageBody, `"`, "", -1)

		ok := validate(prompt)

		if !ok {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   "good try, no funny business, only ASCII. ML-injection is not allowed 😬"})
			return
		}

		fmt.Printf("User %s requested %s \n", user, prompt)

		path, wasBusy, err := generate(ctx, prompt)

		if wasBusy {

			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID: chatId,
					Text:   "I can only generate one image at a time 🐢, try again later",
				})
			return
		}

		if err != nil {
			fmt.Println("Failed to run the model")

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   string("Something went wrong when running the model 😭 "),
			})

			return
		}

		b.SendMessage(ctx,
			&bot.SendMessageParams{ChatID: chatId,
				Text: fmt.Sprintf("Got it! Generating %s for %s  \n", prompt, user),
			})

		fmt.Println("Success. Sending file: ", path)

		fileContent, _ := os.ReadFile(path)

		params := &bot.SendPhotoParams{
			ChatID:  chatId,
			Photo:   &models.InputFileUpload{Filename: "image.png", Data: bytes.NewReader(fileContent)},
			Caption: fmt.Sprintf("%s by %s", prompt, user),
		}

		b.SendPhoto(ctx, params)
	}

	if strings.HasPrefix(message, "@BotatoideBot") {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Just mention me in a message like this: \n\n @BotatoideBot show me a cat \n\n and I'll generate an image for you!",
		})
	}

}

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

const MAGIC_WORDS = "@BotatoideBot"

var (
	BOT_TOKEN   string
	SCRIPT_PATH string
)

func configure() error {
	err := godotenv.Load()
	var ok bool

	if err != nil {
		fmt.Println("Failed to load .env file, fallback on env vars")
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

type workLockerKey string

type workLocker struct {
	working bool
	mut     sync.RWMutex
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

	lock := &workLocker{working: false}

	ctx = context.WithValue(ctx, workLockerKey("workLocker"), lock)

	fmt.Println("Config loaded. Bot is running")

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

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	lock, ok := ctx.Value(workLockerKey("workLocker")).(*workLocker)

	if !ok {
		log.Fatal("workLocker not found")
	}

	chatId := update.Message.Chat.ID
	message := update.Message.Text
	user := update.Message.From.Username

	if strings.HasPrefix(message, MAGIC_WORDS) {

		prompt, ok := getPrompt(message)

		if ok != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   "Sorry your prompt is somehow invalid 😬"})
			fmt.Println("Invalid prompt from", user)
			return
		}

		fmt.Printf("User %s requested %s \n", user, prompt)

		lock.mut.RLock()
		if lock.working {
			lock.mut.RUnlock()
			b.SendMessage(ctx,
				&bot.SendMessageParams{
					ChatID: chatId,
					Text:   "I can only generate one image at a time 🐢, try again later",
				})
			fmt.Println("User", user, "tried to generate an image while another one was being generated")
			return
		}
		lock.mut.RUnlock()

		b.SendMessage(ctx,
			&bot.SendMessageParams{ChatID: chatId,
				Text: fmt.Sprintf("Got it! Generating %s for %s\n", prompt, user),
			})

		lock.mut.Lock()
		lock.working = true
		lock.mut.Unlock()

		args := []string{"-i", SCRIPT_PATH, prompt}

		cmd := exec.Command("zsh", args...)

		err := cmd.Run()

		lock.mut.Lock()
		lock.working = false
		lock.mut.Unlock()

		if err != nil {
			fmt.Println("Failed to run the model")

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatId,
				Text:   string("Something went wrong when running the model 😭 "),
			})

			return
		}

		folderName := strings.Replace(prompt, " ", "_", -1)

		if len(folderName) > 126 {
			folderName = folderName[:126]
		}

		outputPath := strings.ReplaceAll(SCRIPT_PATH, "/run_sd.sh", "") + "/outputs/txt2img-samples/" + folderName + "/seed_27_00000.png"

		fmt.Println("Success. Sending file: ", outputPath)

		fileContent, _ := os.ReadFile(outputPath)

		params := &bot.SendPhotoParams{
			ChatID:  chatId,
			Photo:   &models.InputFileUpload{Filename: "image.png", Data: bytes.NewReader(fileContent)},
			Caption: fmt.Sprintf("%s by %s", prompt, user),
		}

		b.SendPhoto(ctx, params)

		//docParams := &bot.SendDocumentParams{}

		//b.SendDocument(ctx, docParams)

		return
	}

	if strings.Contains(message, "@BotatoideBot") {

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   "Hi! I'm a 🤖 that generates images from text. Just mention me follow by a prompt, like this: \n\n @BotatoideBot a cat in space \n\n and I'll generate an image for you!",
		})
	}

}

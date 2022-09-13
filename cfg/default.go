package cfg

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN   string
	MODEL_URL   string
	OUTPUT_PATH string
	TOKEN       string
)

const MAX_JOBS = 20

func FromEnv() error {
	err := godotenv.Load()
	var ok bool

	if err != nil {
		log.Println("Failed to load .env file, fallback on env vars")
	}

	BOT_TOKEN, ok = os.LookupEnv("BOT_TOKEN")

	if !ok {
		return fmt.Errorf("BOT_TOKEN not found")
	}

	MODEL_URL, ok = os.LookupEnv("MODEL_URL")

	if !ok {
		return fmt.Errorf("MODEL_URL not found")
	}

	OUTPUT_PATH, ok = os.LookupEnv("OUTPUT_PATH")

	if !ok {
		return fmt.Errorf("OUTPUT_PATH not found")
	}

	TOKEN, _ = os.LookupEnv("TOKEN")

	return nil

}

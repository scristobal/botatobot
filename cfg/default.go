package cfg

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN    string
	BOT_USERNAME string
	MODEL_URL    string
	OUTPUT_PATH  string
)

const MAX_JOBS = 10

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

	BOT_USERNAME, ok = os.LookupEnv("BOT_USERNAME")

	if !ok {
		return fmt.Errorf("BOT_USERNAME not found")
	}

	MODEL_URL, ok = os.LookupEnv("MODEL_URL")

	if !ok {
		return fmt.Errorf("MODEL_URL not found")
	}

	OUTPUT_PATH, ok = os.LookupEnv("OUTPUT_PATH")

	if !ok {
		return fmt.Errorf("OUTPUT_PATH not found")
	}

	return nil

}

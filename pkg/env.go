package pkg

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	LOCAL_PORT        string
	TELEGRAMBOT_TOKEN string
	REPLICATE_URL     string
	REPLICATE_TOKEN   string
	REPLICATE_VERSION string
)

func FromEnv() error {
	err := godotenv.Load()

	var ok bool

	if err != nil {
		log.Println("No .env file found, using environment variables instead.")
	}

	LOCAL_PORT, ok = os.LookupEnv("LOCAL_PORT")

	if !ok {
		log.Println("LOCAL_PORT not found, health http rest will not start.")
	}

	TELEGRAMBOT_TOKEN, ok = os.LookupEnv("BOT_TOKEN")

	if !ok {
		return fmt.Errorf("BOT_TOKEN not found. Talk to @botfather in Telegram and get one")
	}

	REPLICATE_URL, ok = os.LookupEnv("REPLICATE_URL")

	if !ok {
		REPLICATE_URL = "https://api.replicate.com/v1/predictions"
		log.Printf("REPLICATE_URL not found, using default %s\n", REPLICATE_URL)

	}

	REPLICATE_TOKEN, ok = os.LookupEnv("REPLICATE_TOKEN")

	if !ok {
		return fmt.Errorf("REPLICATE_TOKEN not found, get a token or set RUN_LOCAL=true and use your own image generator")
	}

	REPLICATE_VERSION, ok = os.LookupEnv("REPLICATE_VERSION")

	if !ok {
		REPLICATE_VERSION = "39ed52f2a78e934b3ba6e2a89f5b1c712de7dfea535525255b1aa35c5565e08b"
		log.Printf("REPLICATE_VERSION not found, using default %s\n", REPLICATE_VERSION)
	}

	return nil

}

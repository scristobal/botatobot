package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN         string
	MODEL_URL         string
	OUTPUT_PATH       string
	REPLICATE_TOKEN   string
	REPLICATE_VERSION string
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
		log.Println("OUTPUT_PATH not found, files will not be saved locally")
	}

	REPLICATE_TOKEN, ok = os.LookupEnv("REPLICATE_TOKEN")

	if !ok {
		log.Println("REPLICATE_TOKEN not found, don't forget to run your models locally")
	}

	REPLICATE_VERSION, ok = os.LookupEnv("REPLICATE_VERSION")

	if !ok {
		log.Println("REPLICATE_VERSION not found, don't forget to run your models locally")
	}

	return nil

}

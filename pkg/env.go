package botatobot

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	PORT              string
	BOT_TOKEN         string
	MODEL_URL         string
	OUTPUT_PATH       string
	REPLICATE_URL     string
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

	PORT, ok = os.LookupEnv("PORT")

	if !ok {
		log.Println("PORT not found, using default 8080")
		PORT = "8080"
	}

	BOT_TOKEN, ok = os.LookupEnv("BOT_TOKEN")

	if !ok {
		return fmt.Errorf("BOT_TOKEN not found")
	}

	OUTPUT_PATH, ok = os.LookupEnv("OUTPUT_PATH")

	if !ok {
		log.Println("OUTPUT_PATH not found, files will not be saved locally")
	}

	RUN_LOCAL, ok := os.LookupEnv("RUN_LOCAL")

	if !ok {
		log.Println("RUN_LOCAL not found, assuming it is false.")
	}

	if RUN_LOCAL == "true" {

		log.Println("RUN_LOCAL=true, using local model")

		MODEL_URL, ok = os.LookupEnv("MODEL_URL")

		if !ok {
			return fmt.Errorf("MODEL_URL not found")
		}

	} else {

		log.Println("RUN_LOCAL=false, using replicate.com as image generator.")

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
			return fmt.Errorf("REPLICATE_VERSION not found, go to replicate.com and choose a model version")
		}
	}

	return nil

}

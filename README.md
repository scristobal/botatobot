# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

![sample image](https://user-images.githubusercontent.com/9478529/216794269-bedc1fa7-3a46-41aa-8ecd-c31175544d44.jpg)

## Yes, but how?

In a nutshell, Botatobot listens for user requests for images and forwards those request towards a image generator cloud service, waits for the response and sends it back to user.

## Pre-requisites

Botatobot works out of the box in combination with [Replicate.com](https://www.replicate.com), [Telegram](https://www.telegram.com) and [fly.io](https://fly.io/).

## Configure and run Botatobot 🥔

Create an `.env` file like so:

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
REPLICATE_TOKEN=1234567890abdfeghijklmnopqrstuvwxyz
```

The `BOT_TOKEN` you got from the [@BotFather](https://t.me/BotFather), `REPLICATE_TOKEN` you get from [Replicate.com api-tokens](https://replicate.com/signin?next=/account/api-tokens)

### Running Botatobot using Docker

Build the docker image with

```bash
docker build .  -t botatobot
```

and then launch it with

```bash
docker run --env-file .env  botatobot
```

As long as the container is running the bot will answer to requests. To make it available 24/7 you need to deploy to a hosted service.

## Deploy Botatobot

If you want to deploy Botatobot you need a [fly.io](https://fly.io/) account. Create the account and [install the CLI](https://fly.io/docs/hands-on/install-flyctl/). Then simply run

```bash
fly launch
```

to create your deployment follow by

```bash
fly secrets set BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
fly secrets set REPLICATE_TOKEN=1234567890abdfeghijklmnopqrstuvwxyz
```

and finally

```bash
fly deploy
```

## Bonus

### Change the model and version

The model version is controlled by `REPLICATE_VERSION` environment variable.

### Run locally w/o Docker

If you have Go 1.21+ you can build the project with

```bash
go build -o build/botatobot cmd/botatobot/main.go
```

and then run

```bash
./build/botatobot
```

### Build and run using Just

If you have Go 1.21+ and [Just](https://just.systems/) you can use

```bash
just -l
```

to check the available commands.

## Acknowledgments

Botatobot uses the excellent [go-telegram](https://pkg.go.dev/github.com/go-telegram/bot) package to interact with the Telegram API. Go check it out!

# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

![sample image](https://user-images.githubusercontent.com/9478529/216794269-bedc1fa7-3a46-41aa-8ecd-c31175544d44.jpg)

## Yes, but how?

In a nutshell, Botatobot listen for users requests for images and push them in a (bounded) worker queue, the queue gets processed by a server running a Stable Diffusion inside a  either locally or as a service, like replicate.com

## Pre-requisites

To run Stable Diffusion, you have two options:

- Locally, using a [Cog Container](https://github.com/replicate/cog).
- Remotely, using [Replicate.com](https://www.replicate.com).

To run the Telegram bot server you need [Go](https://go.dev/doc/install) and a [Telegram](https://www.telegram.com) account.

### Host a Cog Stable Diffusion server on your machine 🐳

First you need to [install Cog](https://github.com/replicate/cog?tab=readme-ov-file#install), [Docker](https://docs.docker.com/get-docker/) and the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html).

Then clone the pre-configured [cog-stable-diffusion](https://github.com/replicate/cog-stable-diffusion) repository. Follow their instructions in the README to ensure the model is running correctly.

In particular you need to download the weights first

```bash
cog run script/download-weights
```

build your container with

```bash
cog build -t stable-diffusion
```

and run it with

```bash
docker run -d -p 5001:5000 --gpus all stable-diffusion
```

This will download, create and run a container with the stable diffusion image running on port 5001, eg.`http://127.0.0.1:5001/predictions`

### Getting a Replicate.com token and model version

Go to [Replicate.com api-tokens](https://replicate.com/signin?next=/account/api-tokens) and generate your token, keep it safe.

You will also need to choose a Stable Diffusion model version. You can find the available versions here <https://replicate.com/stability-ai/stable-diffusion/versions>

### Getting a Telegram bot token 🤖

1. Talk to [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` to create a new bot
3. Follow the instructions to set a name and username for your bot
4. Copy the token that BotFather gives you

The token looks like this: `123456789:ABCDEF1234567890ABCDEF1234567890ABC`

### Configure and run Botatobot 🥔

#### Configure

##### If running Stable Diffusion locally

Set the following environment variables or create a `.env` file with the following content:

```text
TELEGRAMBOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
MODEL_URL=http://127.0.0.1:5001/predictions
```

The `BOT_TOKEN` you got from the [@BotFather](https://t.me/BotFather), and `MODEL_URL` indicates where the Cog Stable Diffusion is running, most likely a docker container in your local machine.

There is an optional variable `OUTPUT_PATH` that indicates the path where the generated images will be saved.

##### If using replicate.com

You need the Telegram and Replicate tokens and the Stable Diffusion model version, create an `.env` file or make them available in the environment.

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
REPLICATE_TOKEN=1234567890abdfeghijklmnopqrstuvwxyz
REPLICATE_VERSION=a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef
```

The `BOT_TOKEN` you got from the [@BotFather](https://t.me/BotFather),`REPLICATE_TOKEN` and `REPLICATE_VERSION` you get from replicate.com.

Additionally you can set `REPLICATE_URL` to a custom url, and `OUTPUT_PATH` to indicate the path where the generated images will be saved.

#### Build and run

Build with`go build -o build/botatobot cmd/botatobot/main.go` and then run `./build/botatobot`

## Usage

Tell the bot `/help` to let him self explain.

## Notes

In some scenarios, like deploying to Heroku or other platforms you need a http rest health endpoint. Botatobot includes such functionality, to activate it include a `LOCAL_PORT` variable.

## Acknowledgments

Botatobot uses the excellent [go-telegram](https://pkg.go.dev/github.com/go-telegram/bot) package to interact with the Telegram API. Go check it out!

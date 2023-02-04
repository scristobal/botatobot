# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

![ri10voizsol91](https://user-images.githubusercontent.com/9478529/216794269-bedc1fa7-3a46-41aa-8ecd-c31175544d44.jpg)

## Yes, but how?

In a nutshell, Botatobot redirects users request for images to a server running a Stable Diffusion [Cog Container](https://github.com/replicate/cog)

Botatobot uses the excellent [go-telegram](<https://pkg.go.dev/github.com/go-telegram/bot@v0.2.2>) package to interact with the Telegram API.

Lastly, in order not to overflow the server it also implements a (bounded) worker queue to handle requests.

## Pre-requisites

To run Stable Diffusion, you need [Docker](https://docs.docker.com/get-docker/) and the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html).

To run the Telegram bot server you need [Go1.18+](https://go.dev/doc/install)

## Running locally 🏃‍♀️

### Host a Stable Diffusion server 🐳

Run Replicate's stable diffusion image locally with `docker run -d -p 5001:5000 --gpus=all r8.im/stability-ai/stable-diffusion@sha256:a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef`.

This will download, create and run a container with the stable diffusion image running on port 5001, eg.`http://127.0.0.1:5001/predictions`

### Getting a Telegram bot token 🤖

1. Talk to [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` to create a new bot
3. Follow the instructions to set a name and username for your bot
4. Copy the token that BotFather gives you

The token looks like this: `123456789:ABCDEF1234567890ABCDEF1234567890ABC`

### Configure and run Botatobot 🥔

#### Configure

Set the following environment variables or create a `.env` file with the following content:

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
MODEL_URL=http://127.0.0.1:5001/predictions
OUTPUT_PATH=/home/user/pictures
```

The variable `OUTPUT_PATH` is optional, and indicates the path where the generated images will be saved.

#### Build and run

Build with`make build` and then run `./build/botatobot`

## Usage

Tell the bot `/help` to let him self explain.

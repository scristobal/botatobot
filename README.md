# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

## Pre-requisites

You need Docker (to run Stable Diffusion) and Go1.18+ installed on your machine (to run the Telegram bot server)

## Running locally 🏃‍♀️

### Running Stable Diffusion

Run Replicate's stable diffusion image locally with `docker run -d -p 5001:5000 --gpus=all r8.im/stability-ai/stable-diffusion@sha256:a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef`.

This will create and run a container with the stable diffusion image running on port 5001, eg.`http://127.0.0.1:5001/predictions`

### Getting a Telegram bot token

1. Talk to [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` to create a new bot
3. Follow the instructions to set a name and username for your bot
4. Copy the token that BotFather gives you

The token looks like this: `123456789:ABCDEF1234567890ABCDEF1234567890ABC`

### Running Botatobot

#### Configure

Set the following environment variables or create a `.env` file with the following content:

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
MODEL_URL=http://127.0.0.1:5001/predictions
OUTPUT_PATH=/home/user/pictures
```

#### Build and run

The variable `OUTPUT_PATH` is optional, and indicates the path where the generated images will be saved.

Build with`make build` and then run `./build/botatobot`

## Usage

Tell the bot `/help` to let him self explain.

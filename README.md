# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

## Setup

Install this fork of stable diffusion <https://github.com/basujindal/stable-diffusion> and the following env variables:

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
BOT_USERNAME=@example_bot
MODEL_PATH=/home/user/repos/stable-diffusion/
OUTPUT_PATH=/home/user/pictures
```

## Usage

To generate images, let the bot know by telling its name followed by a prompt, eg. `@example_bot a potato in a basket`.

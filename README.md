# Botatobot 🥔

Yet another Telegram bot, this one generates stable diffusion images on request.

## Setup

Run Replicate's stable diffusion image locally with `docker run -d -p 5001:5000 --gpus=all r8.im/stability-ai/stable-diffusion@sha256:a9758cbfbd5f3c2094457d996681af52552901775aa2d6dd0b17fd15df959bef` and set the environment variables:

```text
BOT_TOKEN=123456789:ABCDEF1234567890ABCDEF1234567890ABC
BOT_USERNAME=@example_bot
MODEL_URL=http://127.0.0.1:5001/predictions
OUTPUT_PATH=/home/user/pictures
```

## Usage

To generate images, let the bot know by telling its name followed by a prompt, eg. `@example_bot a potato in a basket`.

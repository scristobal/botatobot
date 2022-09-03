#!/bin/zsh
conda activate ldm
cd $MODEL_PATH
python optimizedSD/optimized_txt2img.py --prompt "$1" --W 512 --H 512  --n_iter 1 --n_samples 5 --ddim_steps 50 --turbo --outdir "$2"

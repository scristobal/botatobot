#!/bin/zsh
conda activate ldm
cd "$1"
python optimizedSD/optimized_txt2img.py --prompt "$2" --W 768 --H 768 --seed 27 --n_iter 1 --n_samples 1 --ddim_steps 50
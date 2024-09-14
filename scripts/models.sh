#!/bin/bash

# create models directory if it doesn't exist
mkdir -p models

# setup virtual environment and dependencies
python3 -m venv .env
source .env/bin/activate
pip install -r requirements.txt

# Download GGUF models
huggingface-cli download MaziyarPanahi/gemma-2b-it-GGUF gemma-2b-it.fp16.gguf --local-dir ./models
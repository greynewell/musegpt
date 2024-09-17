#!/usr/bin/env pwsh

# Create models directory if it doesn't exist
if (-Not (Test-Path -Path "models")) {
    New-Item -ItemType Directory -Path "models" | Out-Null
}

# Set up virtual environment and dependencies
python -m venv .env
& .\.env\Scripts\Activate.ps1
pip install -r requirements.txt

# Download GGUF models
huggingface-cli download MaziyarPanahi/gemma-2b-it-GGUF gemma-2b-it.fp16.gguf --local-dir ./models
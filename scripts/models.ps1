# Create models directory if it doesn't exist
New-Item -ItemType Directory -Force -Path .\models

# Setup virtual environment and dependencies
python -m venv .env
.\.env\Scripts\Activate.ps1
pip install -r requirements.txt

# Download GGUF models
huggingface-cli download MaziyarPanahi/gemma-2b-it-GGUF gemma-2b-it.fp16.gguf --local-dir .\models
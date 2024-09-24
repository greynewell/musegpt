# Create models directory if it doesn't exist
New-Item -ItemType Directory -Force -Path .\models

# Setup virtual environment and dependencies
# Remove .env directory
Remove-Item -Path ".env" -Recurse -Force -ErrorAction SilentlyContinue
python -m venv .env
.\.env\Scripts\Activate.ps1
pip install -r requirements.txt

# Download GGUF models
$MODEL_REPO="MaziyarPanahi/gemma-2b-it-GGUF"
$MODEL_FILE="gemma-2b-it.fp16.gguf"
$MODEL_DIR=".\models"

huggingface-cli download $MODEL_REPO $MODEL_FILE --local-dir $MODEL_DIR
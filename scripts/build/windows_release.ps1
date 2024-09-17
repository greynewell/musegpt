#!/usr/bin/env pwsh

# Wipe out build output from this project
$VST3Path = "$env:ProgramFiles\Common Files\VST3\musegpt.vst3"
if (Test-Path -Path $VST3Path) {
    Remove-Item -Path $VST3Path -Recurse -Force
}

# Create build output directories
New-Item -ItemType Directory -Path 'build\release' -Force

# Download models
& ".\scripts\models.ps1"

# Build llama.cpp server
Push-Location -Path 'llama.cpp'
cmake -S . -B ../build/llama.cpp
$cores = [Environment]::ProcessorCount
cmake --build ../build/llama.cpp --target llama-server -- /m:$cores
Pop-Location

# Build main project
Push-Location -Path 'build'
cmake -S .. -B release
cmake --build release --config Release -- /m:$cores
Pop-Location
#!/bin/bash

# Wipe out build output from this project
rm -rf ~/Library/Audio/Plug-Ins/VST3/musegpt.vst3

# Create build output directories
mkdir -p build/release

# Download models
scripts/models.sh

# build llama.cpp server
cd llama.cpp
cmake -S . -B ../build/llama.cpp -D CMAKE_BUILD_TYPE=Release
cmake --build ../build/llama.cpp -j $(sysctl -n hw.physicalcpu) --target llama-server
cd ..

# build main project
cd build
cmake -S .. -B release
cmake --build release --config Release -j $(sysctl -n hw.physicalcpu)
cd ..
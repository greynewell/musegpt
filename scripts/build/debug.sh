#!/bin/bash

# Wipe out build output from this project
rm -rf ~/Library/Audio/Plug-Ins/VST3/musegpt.vst3

# Create build output directories
mkdir -p build/debug

# Download models
scripts/models.sh

# build llama.cpp server
cd llama.cpp
cmake -S . -B ../build/llama.cpp
cmake --build ../build/llama.cpp -j $(sysctl -n hw.physicalcpu) --target llama-server
cd ..

# build main project
cd build
cmake -S .. -B debug
cmake --build debug --config Debug  -DCMAKE_CXX_COMPILER=/pathto/g++ -DCMAKE_C_COMPILER=/pathto/gcc -j $(sysctl -n hw.physicalcpu)
cd ..
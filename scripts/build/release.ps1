# Wipe out build output from this project
Remove-Item -Path "$env:USERPROFILE\Documents\VST3\musegpt.vst3" -Recurse -Force -ErrorAction SilentlyContinue

# Create build output directories
New-Item -Path "build\release" -ItemType Directory -Force

# Download models
& "$PSScriptRoot\..\models.ps1"

# build llama.cpp server
Push-Location llama.cpp
cmake -S . -G "Visual Studio 17 2022" -DCMAKE_CXX_COMPILER="/c/Program Files (x86)/Microsoft Visual Studio/2022/BuildTools/VC/Tools/MSVC/14.41.34120/bin/HostX86/x86/cl" -DCMAKE_C_COMPILER="/c/Program Files (x86)/Microsoft Visual Studio/2022/BuildTools/VC/Tools/MSVC/14.41.34120/bin/HostX86/x86/cl" -B ..\build\llama.cpp
cmake --build ..\build\llama.cpp -j $env:NUMBER_OF_PROCESSORS --target llama-server --config Release
Pop-Location

# build main project
Push-Location build
cmake -S .. -B release
cmake --build release --config Release -j $env:NUMBER_OF_PROCESSORS
Pop-Location
# Requirements [![GitHub Repo stars](https://img.shields.io/github/stars/greynewell/musegpt?style=social)](https://github.com/greynewell/musegpt/stargazers)

## Operating System

- **macOS:** macOS 10.11 or later
- **Windows:** Windows 10 or later
- **Linux:** Mainstream distributions

## DAW Support

Any DAW that supports VST3 plugins (e.g., Ableton Live, FL Studio, Logic Pro, Pro Tools, etc.)

## Dependencies

- **JUCE:** Audio application framework
- **llama.cpp:** LLM inference library
- **Compiler:**
  - macOS: Clang 6.0 or later
  - Windows: Visual Studio 2022 Build Tools for C++
  - Linux: Clang 6.0 or later
- **Python:** 3.10 or later (for model downloading and processing)
- **CMake:** Version 3.15 or later

## Supported Models

**musegpt** currently supports the following models:

- **gemma-2b-it.fp16.gguf**

Any model compatible with `llama.cpp` should work with **musegpt**. Feel free to experiment with different models to find the best one for your needs.

---

*[Back to Home](index.md)*
# Installation

You can install **musegpt** by downloading the latest binaries from the [Releases](https://github.com/greynewell/musegpt/releases) page.

## Build from Source

If you prefer to build from source, follow these steps:

1. **Clone the repository:**

   ```bash
   git clone --recurse-submodules -j2 https://github.com/greynewell/musegpt.git
   cd musegpt
   ```

2. **Install dependencies:**

   Ensure you have the required dependencies installed. See [Requirements](requirements.md) for details.

3. **Build the project:**

     Run the shell build script:

     ```bash
     ./scripts/build/debug.sh
     ```

     or

     ```bash
     ./scripts/build/release.sh
     ```

4. **Install the plugin:**

   CMake will automatically copy the built VST3, AU, or AAX plugin to your DAW's plugin directory. Example paths for VST3 are:
   
   - **macOS:** `~/Library/Audio/Plug-Ins/VST3/`
   - **Linux:** `~/.vst3/`

---

*[Back to Home](index.md)*
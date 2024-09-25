# Installation [![GitHub Repo stars](https://img.shields.io/github/stars/greynewell/musegpt)](https://github.com/greynewell/musegpt/stargazers)

You can install **musegpt** by downloading the latest binaries from the [Releases](https://github.com/greynewell/musegpt/releases) page.

## Build from Source

If you prefer to build from source, follow these steps:

1. **Clone the repository:**
   For all platforms:

   ```bash
   git clone --recurse-submodules -j2 https://github.com/greynewell/musegpt.git
   cd musegpt
   ```

2. **Install dependencies:**

   Ensure you have the required dependencies installed. See [Requirements](https://musegpt.org/requirements.html) for details.

3. **Build the project:**

     For Unix-based systems (Linux, macOS):

     ```bash
     ./scripts/build/debug.sh
     ```

     or

     ```bash
     ./scripts/build/release.sh
     ```

     For Windows (PowerShell):

     ```bash
     ./scripts/build/debug.ps1
     ```

     or

     ```bash
     ./scripts/build/release.ps1
     ```

     **You may need to run the above commands with administrative privileges on Windows.**

     Each build script will also download the relevant model weights for the inference engine.

4. **Install the plugin:**

   CMake will automatically copy the built VST3, AU, or AAX plugin to your DAW's plugin directory. Example paths for VST3 are:
   
   - **macOS:** `~/Library/Audio/Plug-Ins/VST3/`
   - **Linux:** `~/.vst3/`
   - **Windows:** `%USERPROFILE%\Documents\VST3\`

5. **Run the plugin:**

   Start your DAW and you should see **musegpt** in your plugin list. For more detailed instructions on how to use **musegpt**, see the [Usage](https://musegpt.org/usage.html) section of the documentation.

6. **Clean the project (if needed):**

   For Unix-based systems (Linux, macOS):
   ```bash
   ./scripts/clean.sh
   ```

   For Windows (PowerShell):
   ```bash
   ./scripts/clean.ps1
   ```

---

*[Back to Home](index.md)*
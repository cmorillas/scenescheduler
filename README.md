# Scene Scheduler

**Scene Scheduler** is a professional automation software designed to seamlessly control OBS Studio scenes according to a predefined schedule. It acts as an orchestrator, allowing broadcasters and streamers to manage 24/7 channels, switch between live sources, play videos (VOD), and display web content automatically.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Automated Scene Switching:** Control OBS Studio via WebSockets based on precise timing schedules.
- **Multiple Source Support:** Natively schedule local videos (FFmpeg), images, browser sources, and live media streams.
- **Dynamic HLS Generation:** Supports an optional companion tool [`hls-generator`](https://github.com/cmorillas/hls-generator) (C++) to process video streams into HLS without requiring a heavy, system-wide FFmpeg installation.
- **Web Interface:** Control and monitor the schedule remotely via the built-in HTTP server.
- **Desktop GUI:** Built with [Fyne](https://fyne.io/) for cross-platform desktop management.
- **Event-Driven Architecture:** Highly modular backend using an internal event bus for maximum stability.

## Prerequisites

Before running the project, assure you have the following installed:

1. **[Go](https://go.dev/doc/install)** (1.20 or later)
2. **[OBS Studio](https://obsproject.com/download)** (version 28.0+, requires obs-websocket v5)

## Installation & Build Instructions

1. **Clone the repository:**
   ```bash
   git clone https://github.com/cmorillas/scenescheduler.git
   cd scenescheduler
   ```

2. **Build the Application:**

   **Linux:**
   ```bash
   sudo apt-get install -y pkg-config libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libglx-dev libgl1-mesa-dev libxxf86vm-dev libvpx-dev libopus-dev libasound2-dev
   go build -ldflags="-s -w" -o build/scenescheduler .
   ```

   **macOS:**
   ```bash
   brew install pkg-config libvpx opus
   go build -ldflags="-s -w" -o build/scenescheduler .
   ```

   **Windows (MSYS2 / UCRT64):**
   ```bash
   pacman -S mingw-w64-ucrt-x86_64-gcc mingw-w64-ucrt-x86_64-go mingw-w64-ucrt-x86_64-pkgconf mingw-w64-ucrt-x86_64-libvpx mingw-w64-ucrt-x86_64-opus
   go build -ldflags="-s -w -extldflags '-static'" -o build/scenescheduler.exe .
   ```
   > The `-extldflags '-static'` flag embeds the GCC runtime into the executable, so it runs on any Windows machine without needing external DLLs. If static libraries are not available, omit this flag and distribute the MinGW DLLs alongside the `.exe`.

3. **Discover Media Devices:**

   List available video and audio devices to find the correct identifiers for `config.json`:
   ```bash
   ./build/scenescheduler --list-devices
   ```
   See the [User Manual](docs/) for a detailed example of the output and how to configure device identifiers.

4. **Configure Settings:**
   - Modify `config.json` to set your OBS WebSocket password, web server credentials, and the media device identifiers from the previous step.

## Usage

1. Open OBS Studio and ensure the **WebSocket Server** is enabled (Tools -> WebSocket Server Settings). Set the port and password to match your `config.json`.
2. Start the Scene Scheduler application:
   ```bash
   ./build/scenescheduler
   ```
3. The GUI will appear. You can define schedules using the `schedule.json` file or via the internal tools.
4. **HLS Generator (Optional):** If you require HLS streaming conversion for source previews, download the companion tool from its own repository: [hls-generator](https://github.com/cmorillas/hls-generator). Place the binary in the same directory as the `scenescheduler` executable.

## Directory Structure

- `/backend`: Core logic (Scheduler, EventBus, Core Logic, OBS Client, Web Server).
- `/frontend`: Web UI assets and public files.

## Contact

For questions, suggestions, or support: **info@scenescheduler.com**

## License

This project is released under the terms specified in the `LICENSE` file. Please also consult `THIRD-PARTY-LICENSES.txt` for details regarding the open-source libraries used within the application.

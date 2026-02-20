# Scene Scheduler

**Scene Scheduler** is a professional automation software designed to seamlessly control OBS Studio scenes according to a predefined schedule. It acts as an orchestrator, allowing broadcasters and streamers to manage 24/7 channels, switch between live sources, play videos (VOD), and display web content automatically.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Automated Scene Switching:** Control OBS Studio via WebSockets based on precise timing schedules.
- **Multiple Source Support:** Natively schedule local videos (FFmpeg), images, browser sources, and live media streams.
- **Dynamic HLS Generation:** Includes a highly optimized, dynamically-linked `hls-generator` tool (C++) to process video streams into HLS without requiring a heavy, system-wide FFmpeg installation.
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
   git clone https://github.com/yourusername/scenescheduler.git
   cd scenescheduler
   ```

2. **Configure Settings:**
   - Copy or modify the `config.json` file. Ensure you define your OBS WebSocket password, web server credentials, and preferred media devices.

3. **Install Dependencies and Build the Application:**
   ```bash
   go mod download
   go build -o scenescheduler
   ```

## Usage

1. Open OBS Studio and ensure the **WebSocket Server** is enabled (Tools -> WebSocket Server Settings). Set the port and password to match your `config.json`.
2. Start the Scene Scheduler application:
   ```bash
   ./scenescheduler
   ```
3. The GUI will appear. You can define schedules using the `schedule.json` file or via the internal tools.
4. **HLS Generator:** If you require HLS streaming conversion, consult `backend/SPECIFICATIONS.md` for instructions on building and using the standalone C++ generator.

## Directory Structure

- `/backend`: Core logic (Scheduler, EventBus, Core Logic, OBS Client, Web Server).
- `/frontend`: Web UI assets and public files.
- `/hls-generator`: C++ standalone tool for converting varied video inputs to HTTP Live Streaming (HLS) formats.

## License

This project is released under the terms specified in the `LICENSE` file. Please also consult `THIRD-PARTY-LICENSES.txt` for details regarding the open-source libraries used within the application.

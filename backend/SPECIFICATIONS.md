# HLS Generator - Technical Specifications Document

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [General Description](#general-description)
3. [System Architecture](#system-architecture)
4. [Technical Specifications](#technical-specifications)
5. [Program Operation](#program-operation)
6. [User Guide](#user-guide)
7. [System Requirements](#system-requirements)
8. [Compilation and Installation](#compilation-and-installation)
9. [Supported Formats](#supported-formats)
10. [Troubleshooting](#troubleshooting)

---

## Executive Summary

**HLS Generator** is a cross-platform (Linux/Windows) command-line application that converts video files and live streams into HLS (HTTP Live Streaming) format. The program uses FFmpeg libraries included in OBS Studio through **dynamic runtime loading**, eliminating system installation dependencies.

**Key Features:**
- Dynamic FFmpeg loading (no system dependencies)
- Support for multiple input sources (files, SRT, RTMP, NDI, RTSP, web browser)
- Automatic generation of HLS playlist (m3u8) and video segments (.ts)
- Self-contained binaries with no external dependencies
- Automatic updates when OBS Studio is updated

**Version:** 2.0.0
**License:** MIT

---

## General Description

### What is HLS?

HLS (HTTP Live Streaming) is a streaming protocol developed by Apple that fragments multimedia content into small downloadable HTTP segments. It enables adaptive streaming, where the client can adjust quality based on available bandwidth.

### What does this program do?

HLS Generator takes a video file or live stream as input and generates:

1. **playlist.m3u8** - HLS playlist that describes available segments
2. **segmentXXX.ts** - Video segments (typically 6 seconds each)

These files can be served by any HTTP web server and played in browsers, mobile devices, or HLS-compatible players.

### Main Advantage: No Dependencies

Unlike other HLS converters that require FFmpeg installed on the system, **HLS Generator dynamically loads FFmpeg libraries from the OBS Studio installation**, which means:

- **No need to install FFmpeg** on the system
- **Automatically updates** when OBS is updated
- **Smaller binaries** and portable (619 KB on Windows)
- **Compatible with any version** of FFmpeg included in OBS

---

## System Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        HLS Generator                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌───────────────────────────────┐   │
│  │   main.cpp   │─────▶│   HLSGenerator                │   │
│  │  (CLI Entry) │      │   (Orchestrator)              │   │
│  └──────────────┘      └───────────────┬───────────────┘   │
│                                         │                    │
│         ┌───────────────────────────────┼───────────────┐   │
│         │                               ▼               │   │
│  ┌──────┴──────────┐          ┌─────────────────────┐  │   │
│  │  OBSDetector    │          │   FFmpegWrapper     │  │   │
│  │  (Library Path) │          │   (Video Processing)│  │   │
│  └─────────────────┘          └──────────┬──────────┘  │   │
│                                           │             │   │
│                      ┌────────────────────┼─────────┐   │   │
│                      │                    ▼         │   │   │
│               ┌──────┴──────────┐    ┌────────────┴───┐│  │
│               │  FFmpegLoader   │    │  StreamInput   ││  │
│               │  (Dylib Loader) │    │  (Abstraction) ││  │
│               └─────────────────┘    └────────┬───────┘│  │
│                                                │        │   │
│     ┌──────────────────────────────────────────┼─────┐ │   │
│     │                                          ▼     │ │   │
│     │  ┌──────────┐  ┌──────────┐  ┌──────────────┐│ │   │
│     │  │FileInput │  │SRTInput  │  │BrowserInput  ││ │   │
│     │  └──────────┘  └──────────┘  └──────────────┘│ │   │
│     │  ┌──────────┐  ┌──────────┐  ┌──────────────┐│ │   │
│     │  │RTMPInput │  │NDIInput  │  │RTSPInput     ││ │   │
│     │  └──────────┘  └──────────┘  └──────────────┘│ │   │
│     └──────────────────────────────────────────────┘ │   │
│                                                        │   │
└────────────────────────────────────────────────────────────┘
           │
           ▼
    ┌──────────────────────────────┐
    │   OBS Studio Installation    │
    │                               │
    │  libavformat.so/dll           │
    │  libavcodec.so/dll            │
    │  libavutil.so/dll             │
    └──────────────────────────────┘
```

### Main Components

#### 1. **main.cpp** - Entry Point
- Parses command-line arguments
- Initializes logging system
- Coordinates main program flow

#### 2. **OBSDetector** - Library Detection
- Automatically detects OBS Studio installation
- Searches standard installation paths (Linux/Windows)
- Fallback: detects system FFmpeg installation
- Returns path to FFmpeg libraries

#### 3. **FFmpegLoader** - Dynamic Loading
- Dynamically loads symbols from FFmpeg libraries
- Uses `dlopen()` on Linux, `LoadLibrary()` on Windows
- Maps FFmpeg functions to function pointers
- Allows compilation without static FFmpeg linking

#### 4. **FFmpegWrapper** - Video Processing
- Wraps FFmpeg API in a simpler C++ interface
- Manages format contexts, codecs, and packets
- Implements HLS conversion logic
- Handles video segmentation and playlist generation

#### 5. **StreamInput** - Input Abstraction
- Unified interface for different input sources
- Allows treating files and streams uniformly
- Specialized implementations per input type

#### 6. **HLSGenerator** - Main Orchestrator
- Coordinates all components
- Manages conversion lifecycle
- Handles errors and resource cleanup

---

## Technical Specifications

### Language and Standards

- **Language:** C++17
- **Build System:** CMake 3.16+
- **Supported Compilers:**
  - GCC 7+ (Linux)
  - Clang 5+ (Linux/macOS)
  - MinGW-w64 (Windows cross-compilation)
  - MSVC 2017+ (Windows native)

### FFmpeg Libraries Used

The program requires access to the following FFmpeg libraries (dynamically loaded):

- **libavformat** - Container and video format handling
- **libavcodec** - Audio and video codecs
- **libavutil** - Utilities and auxiliary functions

**Recommended version:** FFmpeg 4.0+ (included in OBS Studio 28+)

### HLS Output Format

- **Segment Format:** MPEG-TS (.ts)
- **Segment Duration:** 6 seconds (configurable)
- **Playlist Size:** 5 segments (configurable)
- **Video Codec:** H.264 (inherited from source or transcoded)
- **Audio Codec:** AAC (inherited from source or transcoded)
- **Naming Convention:**
  - Playlist: `playlist.m3u8`
  - Segments: `segment000.ts`, `segment001.ts`, etc.

### HLS Configuration (HLSConfig)

```cpp
struct HLSConfig {
    std::string inputFile;        // Input path or URI
    std::string outputDir;        // Output directory
    int segmentDuration = 6;      // Duration of each segment (seconds)
    int playlistSize = 5;         // Number of segments in playlist
};
```

### Supported Platforms

#### Linux
- **Architectures:** x86_64, ARM64
- **Tested Distributions:** Ubuntu 20.04+, Debian 11+, Fedora 35+
- **Runtime Dependencies:** libc, libstdc++, libdl, libpthread

#### Windows
- **Architectures:** x86_64
- **Versions:** Windows 10 1809+, Windows 11
- **Distribution:** Static self-contained binary (.exe)

---

## Program Operation

### Complete Execution Flow

```
1. START
   ├── Parse arguments (input_source, output_directory)
   └── Initialize logger

2. LIBRARY DETECTION
   ├── Search for OBS Studio installation
   │   ├── Linux: /usr/lib/obs-plugins/
   │   └── Windows: C:\Program Files\obs-studio\
   ├── Fallback: search for system FFmpeg
   └── Return path to libraries

3. DYNAMIC FFMPEG LOADING
   ├── Open libavformat, libavcodec, libavutil
   ├── Resolve required function symbols
   └── Verify successful loading

4. INPUT OPENING
   ├── Detect input type by URI
   │   ├── srt:// → SRTInput
   │   ├── rtmp:// → RTMPInput
   │   ├── http(s):// → BrowserInput
   │   └── (none) → FileInput
   ├── Open source with avformat_open_input()
   ├── Analyze streams with avformat_find_stream_info()
   └── Identify main video stream

5. HLS OUTPUT CONFIGURATION
   ├── Create output directory if it doesn't exist
   ├── Configure HLS format with avformat_alloc_output_context2()
   ├── Create output stream with avformat_new_stream()
   ├── Copy codec parameters
   ├── Set HLS options:
   │   ├── hls_time: segment duration
   │   ├── hls_list_size: playlist size
   │   └── hls_segment_filename: naming pattern
   └── Write header with avformat_write_header()

6. VIDEO PROCESSING
   ├── LOOP: while packets available
   │   ├── Read packet with av_read_frame()
   │   ├── Rescale timestamps with av_packet_rescale_ts()
   │   ├── Write packet with av_interleaved_write_frame()
   │   └── Free packet with av_packet_unref()
   └── Write trailer with av_write_trailer()

7. CLEANUP
   ├── Close input context
   ├── Close output context
   ├── Free resources
   └── Unload libraries

8. END
   └── Return status code (0 = success)
```

### HLS Segmentation Process

HLS Generator delegates segmentation to FFmpeg's HLS muxer, which performs:

1. **Keyframe Analysis:** Identifies I-frames (key frames) in the video
2. **Segment Division:** Cuts video into ~6 second segments at keyframes
3. **Playlist Generation:** Creates `playlist.m3u8` with metadata:
   ```
   #EXTM3U
   #EXT-X-VERSION:3
   #EXT-X-TARGETDURATION:6
   #EXT-X-MEDIA-SEQUENCE:0
   #EXTINF:6.000000,
   segment000.ts
   #EXTINF:6.000000,
   segment001.ts
   ...
   ```

### Memory Management

The program uses manual memory management for FFmpeg resources:

- **AVFormatContext:** Freed with `avformat_close_input()` / `avformat_free_context()`
- **AVCodecContext:** Freed with `avcodec_free_context()`
- **AVPacket:** Freed with `av_packet_free()` / `av_packet_unref()`
- **AVFrame:** Freed with `av_frame_free()` / `av_frame_unref()`

All resources are freed in destructors and `cleanup()` functions.

---

## User Guide

### Installation

#### Prerequisites: Install OBS Studio

**Linux (Ubuntu/Debian):**
```bash
sudo add-apt-repository ppa:obsproject/obs-studio
sudo apt update
sudo apt install obs-studio
```

**Windows:**
Download and install from [https://obsproject.com/download](https://obsproject.com/download)

#### Program Download

**Option 1: Pre-compiled Binary (Windows)**
```bash
# Download release
wget https://github.com/user/hls-generator/releases/download/v2.0.0/hls-generator-windows-x64-static.zip
unzip hls-generator-windows-x64-static.zip
```

**Option 2: Compile from Source (Linux/Windows)**
See [Compilation and Installation](#compilation-and-installation) section

### Basic Usage

#### Syntax

```bash
./hls-generator <input_source> <output_directory>
```

#### Parameters

- **input_source:** Video file path or stream URI
- **output_directory:** Directory where HLS files will be generated

### Usage Examples

#### 1. Convert Video File

```bash
./hls-generator video.mp4 /var/www/html/hls
```

**Output:**
```
/var/www/html/hls/
├── playlist.m3u8
├── segment000.ts
├── segment001.ts
├── segment002.ts
└── ...
```

#### 2. SRT Stream

```bash
./hls-generator srt://192.168.1.100:9000 /tmp/hls-output
```

**Use case:** Receive SRT transmission from OBS Studio and convert to HLS in real-time.

#### 3. RTMP Stream

```bash
./hls-generator rtmp://live.twitch.tv/app/stream_key /path/to/output
```

**Use case:** Re-streaming RTMP content to HLS format.

#### 4. IP Camera (RTSP)

```bash
./hls-generator rtsp://192.168.1.50:554/stream /var/www/cameras/cam1
```

**Use case:** Convert IP camera stream to HLS for web viewing.

#### 5. Web Page Capture

```bash
./hls-generator https://example.com /tmp/web-capture
```

**Use case:** Capture and stream dynamic web content (requires WebKit/WebView2).

#### 6. NDI Source

```bash
./hls-generator ndi://OBS_Studio /output/ndi-stream
```

**Use case:** Capture NDI output from OBS or NDI devices.

### Playing Generated HLS

#### Simple Web Server (Python)

```bash
cd /var/www/html/hls
python3 -m http.server 8080
```

Open in browser: `http://localhost:8080/playlist.m3u8`

#### HTML5 Player

```html
<!DOCTYPE html>
<html>
<head>
    <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
</head>
<body>
    <video id="video" controls width="640" height="360"></video>
    <script>
        var video = document.getElementById('video');
        var hls = new Hls();
        hls.loadSource('http://localhost:8080/playlist.m3u8');
        hls.attachMedia(video);
    </script>
</body>
</html>
```

#### VLC Media Player

```bash
vlc http://localhost:8080/playlist.m3u8
```

#### FFplay

```bash
ffplay http://localhost:8080/playlist.m3u8
```

### Advanced Options (Future)

Currently, HLS configuration is hard-coded with default values. Future versions may include:

```bash
# Configure segment duration
./hls-generator video.mp4 /output --segment-duration 10

# Configure bitrate
./hls-generator video.mp4 /output --video-bitrate 2000k

# Configure resolution
./hls-generator video.mp4 /output --resolution 1280x720

# ABR (Adaptive Bitrate) - multiple qualities
./hls-generator video.mp4 /output --abr 720p,480p,360p
```

---

## System Requirements

### For Execution

#### Linux
- **Operating System:** Linux kernel 4.0+ (Ubuntu 20.04+, Debian 11+, Fedora 35+)
- **Architecture:** x86_64 or ARM64
- **RAM:** 512 MB minimum, 2 GB recommended
- **Disk Space:**
  - Binary: ~2 MB
  - Working space: varies by video (approx. 10% of original size)
- **Required Software:**
  - OBS Studio 28.0+ (recommended) or FFmpeg 4.0+
  - For BrowserInput: GTK3, WebKitGTK

#### Windows
- **Operating System:** Windows 10 1809+ or Windows 11
- **Architecture:** x86_64
- **RAM:** 1 GB minimum, 4 GB recommended
- **Disk Space:**
  - Binary: ~600 KB
  - Working space: varies by video
- **Required Software:**
  - OBS Studio 28.0+ (standard installation in C:\Program Files\obs-studio\)
  - For BrowserInput: Microsoft Edge WebView2 Runtime

### For Compilation

#### Linux
```bash
# Build dependencies
sudo apt install \
    cmake \
    g++ \
    libavformat-dev \
    libavcodec-dev \
    libavutil-dev \
    libgtk-3-dev \
    libjsoncpp-dev

# Minimum versions
cmake >= 3.16
g++ >= 7.0 (with C++17 support)
```

#### Windows (Cross-compilation from Linux)
```bash
# Build dependencies
sudo apt install \
    cmake \
    mingw-w64 \
    g++-mingw-w64-x86-64
```

#### Windows (Native with MSYS2)
```bash
# In MSYS2 MinGW64 shell
pacman -S \
    mingw-w64-x86_64-cmake \
    mingw-w64-x86_64-gcc \
    mingw-w64-x86_64-ffmpeg
```

---

## Compilation and Installation

### Linux Compilation

#### 1. Clone Repository

```bash
git clone https://github.com/user/hls-generator.git
cd hls-generator
```

#### 2. Install Dependencies

```bash
sudo apt update
sudo apt install cmake g++ libavformat-dev libavcodec-dev libavutil-dev
```

**Note:** FFmpeg libraries are only needed for compilation (headers), not for runtime.

#### 3. Compile

```bash
mkdir build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

#### 4. Install (Optional)

```bash
sudo cp hls-generator /usr/local/bin/
```

#### 5. Verify

```bash
./hls-generator
# Should display usage message
```

### Cross-Compilation for Windows (from Linux)

#### 1. Install MinGW

```bash
sudo apt install mingw-w64 g++-mingw-w64-x86-64
```

#### 2. Download FFmpeg Headers

```bash
./setup-ffmpeg-headers-windows.sh
```

This script downloads FFmpeg 7.0 headers for Windows.

#### 3. Compile

```bash
./build-windows.sh
```

This script:
- Creates `build-windows/` directory
- Configures CMake for cross-compilation
- Compiles static binary
- Generates executable: `build-windows/hls-generator.exe`

#### 4. Create Distribution Package

```bash
cd build-windows
zip -j ../hls-generator-windows-x64-static.zip hls-generator.exe
```

**Output:** `hls-generator-windows-x64-static.zip` (~619 KB)

#### 5. Test on Windows

Transfer the `.exe` to Windows and run:

```cmd
hls-generator.exe video.mp4 C:\output
```

### Native Compilation on Windows (MSYS2)

#### 1. Install MSYS2

Download and install from [https://www.msys2.org/](https://www.msys2.org/)

#### 2. Install Dependencies

Open **MSYS2 MinGW64** shell:

```bash
pacman -S \
    mingw-w64-x86_64-cmake \
    mingw-w64-x86_64-gcc \
    mingw-w64-x86_64-ffmpeg
```

#### 3. Compile

```bash
mkdir build
cd build
cmake .. -G "MinGW Makefiles" -DCMAKE_BUILD_TYPE=Release
make -j4
```

#### 4. Run

```bash
./hls-generator.exe video.mp4 /c/output
```

### Verify No FFmpeg Dependencies

#### Linux

```bash
ldd ./hls-generator | grep libav
```

**Expected output:** (empty - no libav* dependencies)

#### Windows

```bash
x86_64-w64-mingw32-objdump -p hls-generator.exe | grep -i dll
```

**Expected output:** Only system DLLs (kernel32.dll, msvcrt.dll, etc.)

---

## Supported Formats

### Input Formats

#### Video Files

| Format | Extension | Support |
|---------|-----------|---------|
| MP4 | .mp4 | ✅ Full |
| Matroska | .mkv | ✅ Full |
| WebM | .webm | ✅ Full |
| AVI | .avi | ✅ Full |
| MOV | .mov | ✅ Full |
| FLV | .flv | ✅ Full |
| MTS/M2TS | .mts, .m2ts | ✅ Full |
| MPEG | .mpg, .mpeg | ✅ Full |
| WMV | .wmv | ✅ Full |

#### Streaming Protocols

| Protocol | URI | Status |
|-----------|-----|--------|
| SRT | srt://host:port | ✅ Implemented |
| RTMP | rtmp://server/app/stream | ✅ Implemented |
| RTSP | rtsp://host/path | ✅ Implemented |
| NDI | ndi://source_name | ✅ Implemented |
| HTTP(S) | http(s)://url | ✅ Implemented (Browser) |

### Supported Video Codecs

| Codec | Decoder | Encoder | Notes |
|-------|---------|---------|-------|
| H.264 | ✅ | ✅ | Recommended for HLS |
| H.265 (HEVC) | ✅ | ✅ | Requires client support |
| VP8 | ✅ | ⚠️ | Transcoding to H.264 |
| VP9 | ✅ | ⚠️ | Transcoding to H.264 |
| AV1 | ✅ | ⚠️ | Transcoding to H.264 |
| MPEG-2 | ✅ | ⚠️ | Transcoding to H.264 |
| MPEG-4 | ✅ | ⚠️ | Transcoding to H.264 |

**Note:** The program attempts to copy HLS-compatible codecs without transcoding. Incompatible codecs are automatically transcoded to H.264.

### Supported Audio Codecs

| Codec | Decoder | Encoder | Notes |
|-------|---------|---------|-------|
| AAC | ✅ | ✅ | Recommended for HLS |
| MP3 | ✅ | ✅ | Compatible with HLS |
| Opus | ✅ | ⚠️ | Transcoding to AAC |
| Vorbis | ✅ | ⚠️ | Transcoding to AAC |
| AC-3 | ✅ | ⚠️ | Transcoding to AAC |
| FLAC | ✅ | ⚠️ | Transcoding to AAC |

### Output Format

**Only supported format:**

- **Container:** MPEG-TS (.ts)
- **Playlist:** HLS m3u8 (version 3)
- **Video:** H.264 (copy or transcoding)
- **Audio:** AAC (copy or transcoding)

---

## Troubleshooting

### Common Issues

#### 1. "OBS Studio not found"

**Symptom:**
```
ERROR: Neither OBS Studio nor FFmpeg found in the system
ERROR: Please install OBS Studio (recommended) or FFmpeg
```

**Solutions:**

**Linux:**
```bash
# Verify OBS installation
which obs
dpkg -l | grep obs-studio

# Reinstall OBS
sudo apt install obs-studio

# Alternative: install FFmpeg
sudo apt install ffmpeg libavformat-dev libavcodec-dev libavutil-dev
```

**Windows:**
```cmd
# Verify OBS installation
dir "C:\Program Files\obs-studio"

# Reinstall OBS from https://obsproject.com/download
```

#### 2. "Failed to load FFmpeg libraries"

**Symptom:**
```
ERROR: Failed to load FFmpeg libraries
ERROR: Failed to initialize HLS generator
```

**Possible causes:**
- OBS installed in non-standard location
- Incompatible OBS version (< 28.0)
- Missing FFmpeg libraries

**Solutions:**

**Linux:**
```bash
# Verify FFmpeg libraries in OBS
ls -la /usr/lib/obs-plugins/libavformat.so*
ls -la /usr/lib/x86_64-linux-gnu/libavformat.so*

# Check which libraries are missing
ldd /usr/lib/obs-plugins/libavformat.so

# Reinstall OBS
sudo apt remove obs-studio
sudo apt install obs-studio
```

**Windows:**
```cmd
# Verify FFmpeg DLLs
dir "C:\Program Files\obs-studio\bin\64bit\avformat*.dll"
dir "C:\Program Files\obs-studio\bin\64bit\avcodec*.dll"
dir "C:\Program Files\obs-studio\bin\64bit\avutil*.dll"

# Reinstall OBS
```

#### 3. "Failed to open input file"

**Symptom:**
```
ERROR: Failed to open input file
```

**Possible causes:**
- File doesn't exist or incorrect path
- Insufficient permissions
- Unsupported format
- Corrupted file

**Solutions:**

```bash
# Verify existence
ls -la /path/to/video.mp4

# Verify permissions
chmod 644 /path/to/video.mp4

# Test with FFmpeg directly
ffprobe /path/to/video.mp4

# Verify integrity
ffmpeg -v error -i /path/to/video.mp4 -f null - 2>&1
```

#### 4. "Failed to setup HLS output"

**Symptom:**
```
ERROR: Failed to setup HLS output
```

**Possible causes:**
- Output directory doesn't exist
- Insufficient permissions in output directory
- Insufficient disk space

**Solutions:**

```bash
# Create directory
mkdir -p /path/to/output

# Verify permissions
chmod 755 /path/to/output

# Check disk space
df -h /path/to/output
```

#### 5. Live Stream Not Working

**Symptom:**
```
ERROR: Failed to open input file
# For SRT/RTMP/RTSP streams
```

**Solutions:**

```bash
# Verify connectivity
ping 192.168.1.100

# Test stream with FFmpeg
ffprobe srt://192.168.1.100:9000
ffplay srt://192.168.1.100:9000

# Check firewall
sudo ufw status
sudo ufw allow 9000/udp  # For SRT
```

#### 6. Incorrect Segmentation

**Symptom:** Segments with incorrect duration or malformed playlist

**Possible causes:**
- Video without regular keyframes
- Variable framerate (VFR)
- Incorrect timestamps in source

**Solutions:**

```bash
# Regenerate keyframes with FFmpeg before using HLS Generator
ffmpeg -i input.mp4 -c:v libx264 -g 60 -keyint_min 60 -c:a copy output.mp4
./hls-generator output.mp4 /output

# For live streams, configure OBS to generate keyframes every 2-6 seconds
```

### Logs and Debugging

#### Enable Detailed Logs

Currently there's no verbose option, but you can modify the code:

In [logger.cpp](src/logger.cpp), change log level:

```cpp
// Change from INFO to DEBUG for more details
Logger::setLevel(LogLevel::DEBUG);
```

Recompile:

```bash
cd build
cmake .. -DCMAKE_BUILD_TYPE=Debug
make
```

#### Use Debugger

**Linux:**
```bash
gdb ./hls-generator
(gdb) run video.mp4 /output
(gdb) bt  # Backtrace in case of crash
```

**Valgrind (Memory leaks):**
```bash
valgrind --leak-check=full ./hls-generator video.mp4 /output
```

### Report Bugs

If you find an issue not listed here:

1. **Gather information:**
   ```bash
   # Program version
   ./hls-generator --version  # (if implemented)

   # OBS version
   obs --version

   # FFmpeg version in OBS
   strings /usr/lib/obs-plugins/libavformat.so | grep "Lavf"

   # Operating system
   uname -a
   cat /etc/os-release  # Linux
   ver  # Windows
   ```

2. **Reproduce the problem with specific command**

3. **Capture complete program output**

4. **Open GitHub issue** with all information

---

## Additional Information

### Known Limitations

1. **No optimized real-time streaming:** The program processes video as fast as possible, without rate limiting for live streaming.

2. **No ABR (Adaptive Bitrate):** Only generates one bitrate/resolution. Doesn't create multi-bitrate playlists.

3. **No encryption:** Doesn't support AES-128 or SAMPLE-AES for protected content.

4. **No DVR for live streams:** Doesn't maintain sliding window of old segments.

5. **Experimental BrowserInput:** Browser capture requires additional configuration and may be unstable.

### Future Roadmap

- [ ] Support for ABR (multiple bitrates)
- [ ] AES-128 encryption
- [ ] Embedded HTTP server to serve HLS
- [ ] DVR mode with sliding window
- [ ] JSON configuration file
- [ ] Optional GUI (GTK/Qt)
- [ ] WebRTC as input support

### References

- **HLS Specification:** [RFC 8216](https://datatracker.ietf.org/doc/html/rfc8216)
- **FFmpeg Documentation:** [https://ffmpeg.org/documentation.html](https://ffmpeg.org/documentation.html)
- **OBS Studio:** [https://obsproject.com](https://obsproject.com)
- **HLS.js (Player):** [https://github.com/video-dev/hls.js](https://github.com/video-dev/hls.js)

### Credits

- **Author:** Cesar
- **License:** MIT
- **FFmpeg:** FFmpeg Project (LGPL/GPL)
- **OBS Studio:** OBS Project (GPL)

### Support

- **Issues:** GitHub Issues
- **Documentation:** [README.md](README.md)
- **Build Instructions:** [WINDOWS-BUILD.md](WINDOWS-BUILD.md)

---

**Document Version:** 1.0
**Last Updated:** 2025-10-16

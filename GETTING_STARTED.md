# Wolarc Installation & Complete User Guide 🐺⚡

This guide provides step-by-step instructions for installing, configuring, running, and connecting to your **Wolarc** Autonomous Local RAG AI Node on Android (Termux) or Linux.

> 📖 **Guides & Documentation**:
> - [**System Architecture & Reference Manual (`DOCUMENTATION.md`)**](DOCUMENTATION.md) — Deep architecture details, vector math, offline hotspot networking, and complete 1-to-1 command matrix.
> - [**Future Milestones & System Roadmap (`ROADMAP.md`)**](ROADMAP.md) — Architectural designs for the resilient client-host connection method and official native Flutter cross-platform companion app.

---


## Table of Contents

1. [System Architecture & Requirements](#1-system-architecture--requirements)
2. [Prerequisites Setup](#2-prerequisites-setup)
   - [Android / Termux (Step-by-Step)](#a-android--termux-setup)
   - [Linux (Debian / Ubuntu / Arch / Fedora)](#b-linux-setup)
3. [Acquiring the LLM Engine (`llama-server`)](#3-acquiring-the-llm-engine-llama-server)
4. [Installing Wolarc](#4-installing-wolarc)
   - [Running the Installer](#running-the-installer)
   - [Conflict Resolution & Setup Modes](#conflict-resolution--setup-modes)
   - [Installing the Modern Go TUI (`wola-tui`)](#installing-the-modern-go-tui-wola-tui)
5. [Downloading or Importing Models](#5-downloading-or-importing-models)
   - [Downloading from Model Store (`wola.gt`)](#downloading-from-model-store-wolagt)
   - [Importing Your Own Local GGUF Files](#importing-your-own-local-gguf-files)
6. [Managing Vector Knowledge & Documents (RAG)](#6-managing-vector-knowledge--documents-rag)
7. [Starting & Operating the Node](#7-starting--operating-the-node)
8. [User Interfaces & Interaction](#8-user-interfaces--interaction)
   - [The Modern Go Bubble Tea TUI (`wola.d` / `wola-tui`)](#the-modern-go-bubble-tea-tui-wolad--wola-tui)
   - [Interface Mode Toggle (`wola tui`)](#interface-mode-toggle-wola-tui)
   - [Standalone Zero-Dependency Web Chat UI](#standalone-zero-dependency-web-chat-ui)
   - [Interactive Terminal Chat (`wola.chat`)](#interactive-terminal-chat-wolachat)
   - [Deep Thinking Stream Viewer (`wola.th`)](#deep-thinking-stream-viewer-wolath)
   - [Interactive Command Wizard (`wola-h`)](#interactive-command-wizard-wola-h)
9. [Connecting Remote Devices & Client Apps](#9-connecting-remote-devices--client-apps)
   - [Tunnel URL & Endpoint (`wola.url`)](#tunnel-url--endpoint-wolaurl)
   - [Air-Gapped Offline & Hotspot Pairing (`wola.local`)](#air-gapped-offline--hotspot-pairing-wolalocal--wolahotspot)
   - [Connecting Chatbox](#connecting-chatbox-mobile--desktop)
   - [Connecting Obsidian](#connecting-obsidian-smart-connections--copilot)
   - [Connecting NextChat / Open-WebUI](#connecting-nextchat--open-webui)
   - [Using Python / OpenAI SDK](#using-the-openai-python-sdk)
10. [Performance Benchmarking & Hardware Tuning](#10-performance-benchmarking--hardware-tuning)
11. [Troubleshooting & FAQ](#11-troubleshooting--faq)
12. [Maintenance & Updates](#12-maintenance--updates)

---

## 1. System Architecture & Requirements

Wolarc runs locally on your device without transmitting prompts or embeddings to external cloud APIs:

```
┌────────────────────────────────────────────────────────┐
│                   Remote Client Apps                   │
│         (Chatbox, Obsidian, NextChat, OpenWebUI)       │
└───────────────────────────┬────────────────────────────┘
                            │ Public HTTPS (SSH Tunnel)
                            ▼
              ┌───────────────────────────┐
              │   Agent API Gateway       │
              │  (agent.py / Port 5000)   │
              └──────┬─────────────┬──────┘
                     │             │
        Query Vector │             │ Augmented Prompt
                     ▼             ▼
┌──────────────────┐  ┌──────────────┐  ┌──────────────┐
│  Knowledge Notes │─>│ Embed Server │  │  Chat LLM    │
│  (indexer.py)    │  │ (Port 8081)  │  │ (Port 8080)  │
└──────────────────┘  └──────────────┘  └──────────────┘
```

### Hardware Specifications
- **Operating System**: Android 8.0+ with Termux, or standard Linux (x86_64, aarch64, armv7).
- **RAM**:
  - **2 GB – 3 GB RAM**: Run 1B to 1.5B chat models (e.g., Llama 3.2 1B, Qwen 2.5 1.5B).
  - **4 GB – 6 GB RAM**: Run 2B to 3B chat models (e.g., Gemma 2 2B, Phi-4 Mini).
  - **8 GB+ RAM**: Run 7B to 8B chat models (e.g., Qwen 2.5 7B, Llama 3.1 8B).
- **Storage**: At least 3 GB free disk space (models range from 800 MB to 4 GB).

---

## 2. Prerequisites Setup

### A. Android / Termux Setup

> [!IMPORTANT]
> **Do not install Termux from Google Play Store** (that build is deprecated and broken). Install Termux from [F-Droid](https://f-droid.org/packages/com.termux/) or [GitHub Releases](https://github.com/termux/termux-app/releases).

1. **Grant Storage Permission**:
   Open Termux and run:
   ```bash
   termux-setup-storage
   ```
   Accept the Android popup prompt. This allows Wolarc to read your document notes from `/sdcard/Documents/agent_knowledge` or import GGUF models from `/sdcard/Download`.

2. **Acquire Termux Wake-Lock**:
   To prevent Android from killing background compute when the screen is turned off:
   ```bash
   termux-wake-lock
   ```
   *(Wolarc also acquires this automatically whenever you start services with `wola.r`).*

3. **Install Core System Packages**:
   ```bash
   pkg update -y && pkg install -y python git curl openssh poppler clang cmake make
   ```
   *(Note: `poppler` provides `pdftotext`, enabling Wolarc to index PDF documents).*

---

### B. Linux Setup

On standard Linux distributions:

- **Ubuntu / Debian**:
  ```bash
  sudo apt update && sudo apt install -y python3 python3-pip git curl openssh-client poppler-utils build-essential cmake
  ```
- **Arch Linux**:
  ```bash
  sudo pacman -S python python-pip git curl openssh poppler base-devel cmake
  ```
- **Fedora / RHEL**:
  ```bash
  sudo dnf install -y python3 python3-pip git curl openssh poppler-utils gcc-c++ cmake make
  ```

---

## 3. Acquiring the LLM Engine (`llama-server`)

Wolarc uses `llama-server` from the [`llama.cpp`](https://github.com/ggerganov/llama.cpp) project as its high-speed C/C++ inference engine.

### Automatic Installation (Default — Zero Manual Steps)
The Wolarc installer (`install.sh`) **automatically installs and configures `llama-server` for you**:
- On **Android / Termux**: It automatically runs `pkg install llama-cpp`, installing the official high-performance prebuilt engine in seconds.
- On **Linux**: It automatically queries and downloads the latest official prebuilt binary for your architecture (`x86_64` or `aarch64`).
- **You do not need to compile anything manually.**

### Manual Installation (If Preferred)

#### Method A: Termux Package Manager (Fastest for Android)
```bash
pkg update && pkg install -y llama-cpp
```

#### Method B: Build from Source (Optional for Custom Compiles)
Advanced users who want custom compiler flags (NEON, AVX2, OpenBLAS):
```bash
cd ~
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
cmake -B build -DGGML_OPENMP=OFF
cmake --build build --config Release -j4 --target llama-server

# Symlink to PATH:
# On Termux:
cp build/bin/llama-server $PREFIX/bin/llama-server
# On Linux:
mkdir -p ~/.local/bin && cp build/bin/llama-server ~/.local/bin/llama-server
```

Verify that the binary is available:
```bash
llama-server --version
```

---

## 4. Installing Wolarc

### Running the Installer (One-Command Setup)

You can install Wolarc completely with just a single command in your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

Alternatively, clone the repository and run `install.sh`:

```bash
cd ~
git clone https://github.com/ovrrup/wolarc.git
cd wolarc
bash install.sh
```

The installer performs an automated 6-step setup:
1. **Hardware & Environment Check**: Evaluates total RAM, available storage, and detects whether running on Termux or Linux.
2. **System Dependencies Check**: Verifies `python3`, `git`, `curl`, `openssh`, `poppler`, `llama-server`, and the `go` compiler.
3. **Pre-existing Setup & Conflict Inspection**: Detects running processes, prior configurations, and installed binaries.
4. **Python Dependencies Setup**: Installs `flask`, `flask-cors`, `requests`, `numpy`, `watchdog`, and `gunicorn`.
5. **Configuration Setup**: Initializes `~/.wolarc` and sets up `MODELS_DIR` and `KNOWLEDGE_DIR`.
6. **Compile & Deploy Binaries**: Automatically compiles the high-performance Go Bubble Tea TUI (`wola-tui`) from source and installs all CLI commands into your path (`$PREFIX/bin` or `~/.local/bin`), symlinking `wola` directly to the TUI.

---

### Conflict Resolution & Setup Modes

If you already have a previous Wolarc installation or running processes, `install.sh` presents three installation strategies:

```
How would you like to handle existing files and directories?
  [1] Delete / Clean Slate (Remove interfering files & stop processes)
      Stops running instances, clears stale locks/logs, removes old binaries.
      Note: Downloaded models in ~/models are safely preserved.

  [2] Adjust & Merge (BETA) (Reconcile paths, adapt config & databases)
      BETA FEATURE: Reconciles existing configs, models, and databases.

  [3] Safe Upgrade (Retain existing configuration and proceed)
```

You can also pass setup modes and permissions directly via CLI flags:
- `bash install.sh --clean`: Clean slate re-install (keeps models).
- `bash install.sh --adjust`: Auto-reconcile with custom paths and existing databases.
- `bash install.sh --upgrade`: Non-destructive update.
- `bash install.sh -y`: Unattended non-interactive install.
- `bash install.sh --tui` / `--no-tui`: Explicitly enable or skip the experimental Go Bubble Tea TUI.
- `bash install.sh --llama` / `--no-llama`: Explicitly enable or skip installing the `llama-server` engine.

---

### Modern Go TUI Integration (`wola` / `wola-tui`)

Wolarc includes an experimental modern interface written in Go with Bubble Tea.
> **Note:** The Go TUI is an experimental design that can be removed completely or opted out of during installation (`--no-tui`). If opted out, the installer skips the ~200MB Go compiler, saves bandwidth/storage, and configures `wola` to launch the classic lightweight interactive shell command center (`wola-h`).

When enabled, `install.sh` compiles `wola-tui` and symlinks it to `wola`.

You can launch the entire suite with just:
```bash
wola
```

Subcommands are also supported directly:
- `wola` — Launch Unified Go & Bubble Tea TUI (Dashboard)
- `wola chat` / `wola c` — Jump directly to Interactive RAG Chat
- `wola models` / `wola m` — Open Model Manager and GGUF Importer
- `wola know` / `wola i` — Vector Knowledge Inspector & Semantic Search Tester
- `wola logs` / `wola l` — Multi-Stream Live Log Follower
- `wola status` — Quick non-interactive node telemetry
- `wola start` / `wola r` — Start all services
- `wola stop` / `wola k` — Stop all services
- `wola restart` / `wola rs` — Restart all services
- `wola url` — Display remote Gateway Tunnel URL

---

## 5. Downloading or Importing Models

Wolarc requires two models to function:
1. A **Chat LLM** (`CHAT_MODEL`) for conversation and reasoning.
2. An **Embedding Model** (`EMBED_MODEL`) for vector RAG calculations.

### Downloading from Model Store (`wola.gt`)

Launch the built-in Model Store:
```bash
wola.gt
# or
wola.gh
```

You will see curated, tested models sorted by device capability:
- **Low RAM Tier (1B - 1.5B)**:
  - `Llama 3.2 1B Instruct Q4_K_M` (~800 MB)
  - `Qwen 2.5 1.5B Instruct Q4_K_M` (~1.1 GB)
- **Balanced Tier (2B - 3B)**:
  - `Gemma 2 2B IT Q4_K_M` (~1.6 GB)
  - `Phi-4 Mini 3.8B Q4_K_M` (~2.4 GB)
- **High Capacity Tier (7B - 8B)**:
  - `Qwen 2.5 7B Instruct Q4_K_M` (~4.7 GB)
- **Embedding Models**:
  - `bge-small-en-v1.5 Q8_0` (~70 MB)
  - `nomic-embed-text-v1.5 Q8_0` (~150 MB)

Selecting a model downloads it directly from Hugging Face into `~/models` with live progress bars and automatically configures it in `~/.wolarc`.

---

### Importing Your Own Local GGUF Files

If you already downloaded `.gguf` files on your device (e.g., in `/sdcard/Download` or an external drive), Wolarc provides built-in importing:

1. **Via Model Store**:
   Launch `wola.gt` and select option **[5] Import Local GGUF Model from Device Storage**.
2. **Via Command Line**:
   ```bash
   wola.gt import /sdcard/Download/my-model.gguf
   # or
   wola.m import /sdcard/Download/my-model.gguf
   ```

When importing:
- Wolarc prompts you whether to **[1] Move** the file (saving disk space) or **[2] Copy** the file.
- It asks whether to assign the file as your active **Chat Model** or **Embedding Model** in `~/.wolarc`.

---

## 6. Managing Vector Knowledge & Documents (RAG)

Wolarc's background indexer automatically watches your notes directory and indexes them using vector cosine similarity.

### Default Directories
- **Knowledge Base (`KNOWLEDGE_DIR`)**:
  - Termux: `~/storage/shared/Documents/agent_knowledge` (maps to `/sdcard/Documents/agent_knowledge`).
  - Linux: `~/Documents/agent_knowledge`.
- **Chat Conversations (`CHATS_DIR`)**:
  - Termux: `~/storage/shared/Documents/agent_chats` (maps to `/sdcard/Documents/agent_chats`).
  - Linux: `~/Documents/agent_chats`.

### Supported File Formats (Zero Limitations)
You can drop virtually any file type into your knowledge directory:
- 📝 **Notes & Documents**: `.md`, `.txt`, `.pdf`, `.docx` (Word), `.epub` (E-books), `.rtf`, `.odt`, `.html`, `.tex`, `.org`
- 🎵 **Audio Recordings & Music**: `.mp3`, `.wav`, `.m4a`, `.ogg`, `.flac`, `.opus`, `.aac`, `.wma` (extracts metadata, duration, artist, lyrics, and companion transcripts `.lrc`/`.srt`/`.txt`)
- 🎬 **Video Media**: `.mp4`, `.mkv`, `.mov`, `.avi`, `.webm`, `.flv`, `.wmv` (extracts container metadata, duration, resolution, chapters, and embedded/companion subtitles `.srt`/`.vtt`)
- 🖼️ **Photos, Diagrams & Graphics**: `.jpg`, `.jpeg`, `.png`, `.webp`, `.svg`, `.gif`, `.bmp` (extracts EXIF camera tags, date taken, GPS coordinates, dimensions, captions, and Tesseract OCR visual text)
- 📊 **Office Sheets & Slides**: `.pptx` (slide text), `.xlsx` (tables & cell strings), `.csv`, `.tsv`, `.ods`
- 💻 **Source Code & Scripts**: `.py`, `.js`, `.ts`, `.html`, `.css`, `.json`, `.yaml`, `.toml`, `.sql`, `.sh`, `.c`, `.cpp`, `.rs`, `.go`, `.java`, etc.
- 💬 **Conversation Logs**: `.md`, `.chat` files from past interactions are automatically chunked by question-and-answer pairs.
- 📦 **Universal Fallback**: Any binary or data file is parsed for printable text sequences and indexed.

### Autonomous Intelligence & Long-Term Conversation Memory
Unlike basic RAG systems that merely quote raw document excerpts or say "The notes do not say this", Wolarc's AI operates with **foundational autonomy and intelligence**:
1. **Intelligence First**: The model thinks critically and answers using its own vast pre-trained capabilities and logic.
2. **Conversation Learning**: All past dialogues from the Web UI, Go TUI, CLI `wola.chat`, or API are saved into `CHATS_DIR` and asynchronously indexed. When you talk to the agent in subsequent sessions, it recalls your past decisions, preferences, and personal facts.
3. **Dual Context Blending**: Context is cleanly categorized into `[Past Conversation Memory & User Context]` and `[Personal Knowledge & Memory Context]`.
4. **Natural Synthesis**: The model seamlessly weaves your facts, code snippets, and preferences into responses without robotic disclaimers.

### Indexing CLI Commands (`wola.i`)
```bash
# Check indexer status, total chunks, and document count with media icons
wola.i status

# Manually add a file or folder and index it immediately
wola.i add ~/storage/shared/Documents/meeting_recording.mp3
wola.i add /path/to/media_or_docs_folder/

# Test semantic search query directly from terminal
wola.i search "What are the project deliverables?"
```

Whenever you chat with Wolarc (via TUI, Web UI, terminal, or remote clients), it semantically discovers matching context and integrates it into the AI's personal memory.

---

## 7. Starting & Operating the Node

### Start the Node (`wola.r`)
```bash
wola.r
```

This launches the full background stack:
1. Embedding Engine on port `8081`.
2. Chat Engine on port `8080`.
3. Document Indexer Daemon (`indexer.py`).
4. OpenAI-Compatible API Gateway on port `5000`.
5. Resilient Outbound Tunnel (`localhost.run` -> `serveo.net` -> `pinggy.io`).
6. Acquires Android CPU wake-lock.

### Check Status (`wolarc status`)
```bash
wolarc status
```
Displays running PIDs and active listening ports.

### Stop the Node (`wola.k`)
```bash
wola.k
```
Cleanly halts all engines, terminates tunnels, clears lock files, and releases the wake-lock.

### Restart the Node (`wola.rs`)
```bash
wola.rs
```
Performs a stop, wait, and fresh start cycle.

---

## 8. User Interfaces & Interaction

### The Modern Go Bubble Tea TUI (`wola.d` / `wola-tui`)

Launch the telemetry dashboard:
```bash
wola.d
# or
wola-tui
```

#### Dashboard Navigation:
- `[1]`: **Telemetry Dashboard**:
  - Live Unicode sparklines (` ▂▃▄▅▆▇█`) for token speed and system RAM.
  - Context window and memory utilization gradient bars.
  - Active chat/embed model badges and RAG statistics.
  - Outbound tunnel radar indicator.
- `[2]`: **Models Browser**:
  - Browse local models in `~/models`.
  - Assign active roles on the fly.
- `[3]`: **Knowledge Catalog**:
  - Review indexed documents, chunk sizes, and trigger manual re-indexes (`[i]`).
- `[4]`: **Live Logs Viewer**:
  - Stream logs for chat, embed, agent, tunnel, or thinking.
  - Press `/` to open the live keyword filter drawer.
- `[t]`: **Parameter Tuner Drawer**:
  - Adjust Temperature, Top-P, Threads, and Context Window in real time.
- `Ctrl+K`: **Fuzzy Command Palette**:
  - Jump to views or execute actions instantly.
- `?`: **Help Overlay Cheat Sheet**.
- `q`: Exit dashboard.

---

### Interface Mode Toggle (`wola tui`)

Wolarc supports two distinct operational modes for terminal interaction:
1. **Modern Go Bubble Tea TUI** (Default): Full-screen multi-tab dashboard, interactive chat, model manager, and live log followers.
2. **Classic Unix CLI Suite**: Fast, modular terminal scripts (`wola.r`, `wola.k`, `wola.chat`, `wola.bench`, `wola-h`).

You can toggle your default launch mode at any time:
```bash
# Enable Go Bubble Tea TUI as default
wola tui on

# Disable TUI and use classic CLI menu suite as default
wola tui off

# Check current mode status
wola tui status
```
When TUI mode is disabled, running `wola` opens the classic `wola-h` menu wizard. You can still launch the TUI explicitly at any time by running `wola-tui`.

---

### Standalone Zero-Dependency Web Chat UI

When Wolarc is running, it hosts an embedded, 100% air-gapped Web Chat UI directly on port `5000`. You can access it from any browser on your device or over your local network:
- **Local Browser**: `http://localhost:5000/` or `http://127.0.0.1:5000/`
- **Wi-Fi LAN**: `http://192.168.1.X:5000/` (check with `wola.local`)
- **Android Hotspot**: `http://192.168.43.1:5000/`

#### Key Web UI Highlights:
- **Zero External Dependencies**: Works completely offline with zero CDN dependencies, tracking scripts, or remote web fonts.
- **SSE Token Streaming**: Live token-by-token response streaming with elapsed latency and token velocity tracking.
- **Deep Thinking Support**: Automatically formats reasoning chains from models like DeepSeek-R1 inside collapsible `<think>` process cards.
- **Syntax Highlighting & Copy**: Clean markdown rendering with one-click code copy buttons.
- **Connection Helper Modals**: Provides quick copy-paste configuration commands for curl, Python SDK, Obsidian, and desktop chat apps.

---

### Interactive Terminal Chat (`wola.chat`)

Chat directly with your local AI node in the terminal:
```bash
wola.chat
# or
wola.c
```

#### Features:
- Real-time token streaming.
- Automatic RAG citation boxes displaying matched notes, file paths, and cosine similarity percentages.
- In-chat slash commands:
  - `/multi`: Enter multi-line input mode.
  - `/reset`: Clear chat context history.
  - `/think`: Toggle display of reasoning chains (`<think>`).
  - `/temp <val>`: Adjust sampling temperature (e.g., `/temp 0.2`).
  - `/exit`: Exit chat.

---

### Model Thinking & Reasoning Engine (`wola think` / `wola.th`)

For reasoning models like DeepSeek-R1 or Qwen-2.5:
```bash
# Toggle reasoning tokens on or off to save battery, RAM, and time
wola think on          # Models emit full reasoning thought process (<think>)
wola think off         # Suppress thinking tokens for direct, low-latency answers
wola think status      # Inspect current thinking setting

# View reasoning inner monologue in real time
wola.th
```

You can also toggle thinking dynamically at runtime:
- **Web Chat UI**: Click the `🧠 Thinking: ON` / `⚡ Thinking: OFF` button in the top navigation bar.
- **Go TUI**: Open Command Palette (`Ctrl+K`) and select `Toggle Thinking Mode`.
- **CLI Chat (`wola.chat`)**: Type `/think` directly in conversation.

---

### Interactive Command Wizard (`wola-h`)

```bash
wola-h
```
Displays a clean command center menu with shortcuts to run any operation without having to remember CLI commands.

---

## 9. Connecting Remote Devices & Client Apps

Wolarc exposes a standard OpenAI-compatible API at `/v1`. You can connect client apps from your phone, laptop, or tablet.

### Tunnel URL & Endpoint (`wola.url`)

Run:
```bash
wola.url
```

Output includes:
- **Public Tunnel URL**: e.g., `https://random-id.lhr.life`
- **OpenAI Endpoint**: `https://random-id.lhr.life/v1`
- **Local LAN Endpoint**: `http://192.168.1.X:5000/v1`
- **Ping Latency**: Confirms round-trip connectivity.
- **ASCII QR Code**: Scan with your phone camera to open or copy the URL.
- **Automatic Clipboard Sync**: The URL is copied to your clipboard automatically.

---

### Air-Gapped Offline & Hotspot Pairing (`wola.local` / `wola.hotspot`)

When you are off the grid, traveling, or without internet access, Wolarc can operate entirely air-gapped across local network interfaces:

```bash
# Display all detected local addresses, Web UI link, and terminal QR code
wola.local
# or
wola.hotspot

# Display standalone QR code for scanning from mobile cameras
wola.local qr

# Copy local URL to clipboard
wola.local copy
```

#### Supported Offline Scenarios:
1. **Local Wi-Fi Network**: Any computer or mobile phone on the same Wi-Fi router can access Wolarc at `http://192.168.1.X:5000/`.
2. **Android Portable Hotspot**: Turn on your Android phone's Hotspot (no cellular data needed). Connect your laptop or tablet to the hotspot and access Wolarc at `http://192.168.43.1:5000/`.
3. **USB Reverse Tethering / RNDIS**: Connect a USB cable directly from your Android phone to your PC and browse to `http://192.168.42.129:5000/`.

---

### Connecting Chatbox (Mobile & Desktop)

[Chatbox](https://chatboxai.app/) is a popular cross-platform client for iOS, Android, macOS, and Windows.

1. Open **Chatbox Settings** -> **Model Provider**.
2. Select **OpenAI API Compatible** (or Custom Provider).
3. Set **API Host / Base URL**:
   `https://<your-tunnel-url>/v1` (or local `http://192.168.1.X:5000/v1`)
4. Set **API Key**: Enter any dummy text (e.g., `wola-local`).
5. Set **Model**: Enter your active model name or leave default.
6. Click **Save** and start chatting!

---

### Connecting Obsidian (Smart Connections / Copilot)

You can turn Wolarc into the private AI brain for your Obsidian vault:

1. Install the **Copilot** or **Smart Connections** community plugin in Obsidian.
2. In the plugin settings, choose **Custom OpenAI-compatible Endpoint**.
3. Set **Base URL**: `https://<your-tunnel-url>/v1` (or `http://127.0.0.1:5000/v1` if running Obsidian on the same machine).
4. Set **API Key**: `not-needed`.
5. Now you can query your Obsidian notes using your local Termux node!

---

### Connecting NextChat / Open-WebUI

- **NextChat (ChatGPT-Next-Web)**:
  - Settings -> Model Provider -> Custom
  - Endpoint: `https://<your-tunnel-url>/v1`
  - API Key: `wola`
- **Open-WebUI**:
  - Admin Panel -> Settings -> Connections -> Add OpenAI API
  - URL: `https://<your-tunnel-url>/v1`
  - Key: `wola`

---

### Using the OpenAI Python SDK

You can connect to Wolarc using any standard Python script:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://<your-tunnel-url>/v1",  # Or http://127.0.0.1:5000/v1
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="local-model",
    messages=[
        {"role": "user", "content": "Summarize my notes on the quarterly budget."}
    ],
    stream=True
)

for chunk in response:
    content = chunk.choices[0].delta.content
    if content:
        print(content, end="", flush=True)
print()
```

---

## 10. Performance Benchmarking & Hardware Tuning

### Automated Hardware Profiler (`wola.bench`)

To discover the optimal CPU thread count for your device and prevent thermal throttling:
```bash
wola.bench
```

- Tests prompt processing (PP) and token generation (TG) across 1 to 8 threads.
- Recommends the highest tokens/second configuration that avoids thermal throttling.
- Prompts you to apply the optimal `CHAT_THREADS` setting directly to `~/.wolarc`.

### Manual Parameter Tuning (`wola.t`)

View or adjust parameters in `~/.wolarc`:
```bash
# View all settings
wola.t

# Tune parameters
wola.t TEMPERATURE 0.3
wola.t CHAT_THREADS 4
wola.t FLASH_ATTENTION auto
wola.t MLOCK true
```

---

## 11. Troubleshooting & FAQ

### 1. `llama-server: command not found`
- **Cause**: The `llama-server` binary is not in your system `$PATH`.
- **Fix**: Check where you compiled `llama.cpp` and set it in `~/.wolarc`:
  ```bash
  wola.t LLAMA_BIN /path/to/llama.cpp/build/bin/llama-server
  ```
  Or copy it to `$PREFIX/bin`:
  ```bash
  cp ~/llama.cpp/build/bin/llama-server $PREFIX/bin/
  ```

### 2. Service failed to start / Process killed (OOM)
- **Cause**: The model requires more RAM than your device has available.
- **Fix**:
  - Download a smaller quantized model using `wola.gt` (e.g., Llama 3.2 1B or Qwen 2.5 1.5B with `Q4_K_M`).
  - Reduce context window: `wola.t CHAT_CTX 2048`.
  - Close heavy background apps on Android.

### 3. Tunnel URL not connecting / Delayed responses
- **Cause**: Outbound SSH tunnel disconnected or network latency.
- **Fix**: Run `wola.url reconnect` to re-negotiate an encrypted tunnel with a fresh provider.

### 4. Vector indexer not processing PDF files
- **Cause**: Missing `pdftotext` utility.
- **Fix**: Run `pkg install poppler` (on Termux) or `sudo apt install poppler-utils` (on Linux).

### 5. Port already in use / Stale lock
- **Cause**: A previous process did not exit cleanly.
- **Fix**:
  ```bash
  wola.k
  ```
  If necessary:
  ```bash
  pkill -9 -f llama-server; pkill -9 -f agent.py; rm -f ~/.wola.lock
  ```

---

## 12. Maintenance & Updates

### Checking for Software Updates (`wola.u`)
```bash
wola.u
```
Compares your local repository with the latest GitHub release, pulls updates, reconciles dependencies, and updates binaries while preserving your custom configurations in `~/.wolarc`.

### Comprehensive Command & Discrete Operations Reference
For an exhaustive reference covering every CLI command syntax, configuration variable, and the discrete multi-level cleanup & self-destruct command (`wola.purge`), consult:
- [`COMMANDS.md`](file:///data/data/com.termux/files/home/wola/COMMANDS.md)

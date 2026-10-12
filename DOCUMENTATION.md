# Wolarc System Architecture & Complete Reference Manual 🐺⚡

> **Wolarc** is an autonomous, fully private, local RAG (Retrieval-Augmented Generation) AI node engineered for Android (via Termux) and Linux platforms. It orchestrates high-performance GGUF quantized models via `llama.cpp`, maintains an automated vector knowledge indexer over local files and conversations, exposes an OpenAI-compatible HTTP/SSE API gateway, provides an air-gapped zero-dependency Web UI, and establishes encrypted remote tunnels for off-network client pairing.
>
> 📚 **Official GitHub Wiki**: Browse the complete online documentation encyclopedia at [**github.com/ovrrup/wolarc/wiki**](https://github.com/ovrrup/wolarc/wiki) (or offline local copy in [`wiki/`](wiki/Home.md)).

---


## Table of Contents

1. [Executive Summary & System Philosophy](#1-executive-summary--system-philosophy)
2. [Complete System Architecture & Dataflow](#2-complete-system-architecture--dataflow)
   - [Architectural Topology Diagram](#architectural-topology-diagram)
   - [Core Subsystem Components](#core-subsystem-components)
3. [The Strict 1-to-1 Technical vs. Casual Command Model](#3-the-strict-1-to-1-technical-vs-casual-command-model)
   - [Command Pairing Rationale](#command-pairing-rationale)
   - [Complete Command Matrix](#complete-command-matrix)
   - [Unified Subcommand Dispatcher (`wola` / `wolarc`)](#unified-subcommand-dispatcher-wola--wolarc)
4. [Exhaustive Command Reference Manual](#4-exhaustive-command-reference-manual)
   - [Lifecycle & Process Management (`wola.start` / `wola.r`, `wola.stop` / `wola.k`, `wola.restart` / `wola.rs`, `wola.check` / `wola.status`)](#lifecycle--process-management)
   - [Telemetry Dashboard & Interactive TUI (`wola.dashboard` / `wola.d`, `wola.ui` / `wola.tui`)](#telemetry-dashboard--interactive-tui)
   - [Conversational Clients (`wola.chat` / `wola.c`)](#conversational-clients)
   - [Model Management & Store (`wola.models` / `wola.m`, `wola.store` / `wola.gt`)](#model-management--store)
   - [Vector Knowledge Base & RAG Indexer (`wola.knowledge` / `wola.i`)](#vector-knowledge-base--rag-indexer)
   - [Hardware Profiling & Auto-Tuning (`wola.bench` / `wola.b`)](#hardware-profiling--auto-tuning)
   - [Diagnostics & Live Logging (`wola.log` / `wola.l`, `wola.think` / `wola.th`)](#diagnostics--live-logging)
   - [Gateways & Network Connectivity (`wola.link` / `wola.url`, `wola.hotspot` / `wola.local`, `wola.tailscale` / `wola.ts`)](#gateways--network-connectivity)
   - [Configuration Parameter Tuning (`wola.config` / `wola.t`)](#configuration-parameter-tuning)
   - [Maintenance & Updates (`wola.update` / `wola.u`)](#maintenance--updates)
   - [Interactive Command Center (`wola.help` / `wola.h`)](#interactive-command-center)
   - [System Cleanup & Deletion Engine (`wola.uninstall` / `wola.purge`)](#system-cleanup--deletion-engine)
5. [Universal RAG Pipeline & Semantic Memory](#5-universal-rag-pipeline--semantic-memory)
   - [Multi-Format Ingestion Engine](#multi-format-ingestion-engine)
   - [Vector Chunking & Embedding Generation](#vector-chunking--embedding-generation)
   - [Vector Database Architecture (`knowledge_db.json`)](#vector-database-architecture)
   - [Cosine Similarity Scoring & Dynamic Thresholding](#cosine-similarity-scoring--dynamic-thresholding)
   - [Memory Blending: Base Intelligence vs. Personal Context](#memory-blending-base-intelligence-vs-personal-context)
6. [Offline Air-Gapped Networking, Tailscale Mesh & Remote Tunneling](#6-offline-air-gapped-networking-tailscale-mesh--remote-tunneling)
   - [Local Wi-Fi LAN Discovery](#local-wi-fi-lan-discovery)
   - [Android Portable Hotspot Pairing](#android-portable-hotspot-pairing)
   - [USB Reverse Tethering / RNDIS](#usb-reverse-tethering--rndis)
   - [Tailscale WireGuard Mesh VPN Integration](#tailscale-wireguard-mesh-vpn-integration)
   - [Cloudflare Quick Tunnels](#cloudflare-quick-tunnels)
   - [Autonomous Tunnel Watchdog Daemon (`wola-tunnel-daemon`)](#autonomous-tunnel-watchdog-daemon)
   - [Multi-Route Reachability & Latency Benchmark (`wola.link test`)](#multi-route-reachability--latency-benchmark)
7. [Inference Engine & Deep Reasoning Controls](#7-inference-engine--deep-reasoning-controls)
   - [Deep Thinking Mechanism (`<think>`)](#deep-thinking-mechanism-think)
   - [Dual Engine Isolation (Chat vs. Embed)](#dual-engine-isolation-chat-vs-embed)
8. [Configuration Knobs & Environment Variables](#8-configuration-knobs--environment-variables)
9. [Third-Party Client App Integration Guide](#9-third-party-client-app-integration-guide)
10. [Daemon Lifecycle & Operating System Integration](#10-daemon-lifecycle--operating-system-integration)
11. [Future Milestones & System Roadmap](#11-future-milestones--system-roadmap)
    - [Milestone 1: Reliable & Effective Client-Host Connection Method](#milestone-1-reliable--effective-client-host-connection-method)
    - [Milestone 2: Native Cross-Platform Flutter Companion App](#milestone-2-native-cross-platform-flutter-companion-app)
    - [Phased Delivery Schedule](#phased-delivery-schedule)

---

## 1. Executive Summary & System Philosophy

Wolarc was built around three guiding engineering principles:

1. **Total Sovereignty & Privacy**: Zero telemetric phoning home, zero reliance on external cloud inference APIs, and 100% offline capability. Your data, document vectors, prompts, and thinking traces never leave your physical device unless you explicitly connect through your private tunnel.
2. **Resource-Conscious Mobile Efficiency**: Tailored specifically for ARM64 mobile hardware (Android Termux) and resource-constrained edge Linux boxes. It enforces CPU wake-locks, allocates threads intelligently across LITTLE and big core clusters, implements Flash Attention to minimize memory bandwidth, and isolates embedding compute from chat inference.
3. **Ergonomic Simplicity (The 1-to-1 Command Model)**: Users should never be confused by a tangle of 4 redundant names for the same action or cryptic abbreviations without friendly counterparts. Every single capability in Wolarc is mapped strictly to **one Technical name** (for fast terminal typing and shell scripts) and **one Casual name** (for intuitive, self-explanatory human interaction).

---

## 2. Complete System Architecture & Dataflow

### Architectural Topology Diagram

```
                     ┌────────────────────────────────────────────────────────┐
                     │                   Remote Client Apps                   │
                     │         (Chatbox, Obsidian, NextChat, OpenWebUI)       │
                     └───────────────────────────┬────────────────────────────┘
                                                 │
                                 ┌───────────────┴───────────────┐
                                 │                               │
                       HTTPS SSH Tunnel              Local Offline Network
                     (localhost.run / serveo)      (Wi-Fi / AP Hotspot / USB)
                                 │                               │
                                 └───────────────┬───────────────┘
                                                 │
                                                 ▼
               ┌──────────────────────────────────────────────────────────────────┐
               │              Agent API Gateway (agent.py / Port 5000)            │
               │  • Flask + Gunicorn WSGI Server                                  │
               │  • Standalone Air-Gapped Zero-Dependency Web Chat UI (GET /)     │
               │  • OpenAI-Compatible SSE Streaming (/v1/chat/completions)        │
               │  • Network Discovery Interface (/v1/local_info)                  │
               │  • Dynamic Cognitive Reasoning Filter (<think> parser)           │
               └───────────────────┬──────────────────────────┬───────────────────┘
                                   │                          │
                 Cosine Retrieval  │                          │ Augmented Prompt
                 & Semantic Memory │                          │ with Injected Context
                                   ▼                          ▼
      ┌─────────────────────────────────────────┐  ┌─────────────────────────────────┐
      │   Embedding Server (Port 8081)          │  │   Primary Chat LLM (Port 8080)  │
      │   llama-server --embedding              │  │   llama-server (Chat Engine)    │
      │   Dedicated CPU core allocation         │  │   Parallel slots, Flash-Attn    │
      └────────────────────▲────────────────────┘  └─────────────────────────────────┘
                           │
             Incremental Vectorization
                           │
      ┌─────────────────────────────────────────┐
      │   Universal File Indexer (indexer.py)   │
      │   • Watchdog background filesystem loop │
      │   • Deduplication (SHA-256 caching)     │
      │   • Audio, Video, PDF, Docs, Chats      │
      └────────────────────▲────────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
┌───────────────┴──────────────┐   ┌──────────┴───────────────────┐
│ User Knowledge Directory     │   │ Saved Dialogue Memories       │
│ (~/.../agent_knowledge)      │   │ (~/.../agent_chats)          │
│ Notes, PDFs, Media, Code     │   │ Auto-logged markdown turns   │
└──────────────────────────────┘   └──────────────────────────────┘
```

### Core Subsystem Components

1. **Primary Chat Inference Engine (`llama-server` on Port 8080)**:
   - Hosts the primary Large Language Model (e.g. Qwen 2.5, Llama 3.2, Gemma 2, Phi-4).
   - Serves an OpenAI-compatible HTTP completion API with token-by-token Server-Sent Events (SSE).
   - Configured with optimal context window allocation (`CHAT_CTX`), multi-slot parallel handling (`CHAT_PARALLEL`), memory locking (`MLOCK`), and Flash Attention (`FLASH_ATTENTION`).

2. **Dedicated Embedding Engine (`llama-server` on Port 8081)**:
   - Hosts a specialized small vector embedding model (e.g. `embeddinggemma-300M`, `bge-small-en-v1.5`, `nomic-embed-text-v1.5`).
   - Runs in dedicated embedding mode (`--embedding`), isolating high-frequency vector operations from the chat context to eliminate KV-cache thrashing.

3. **Universal Background Indexer Daemon (`indexer.py`)**:
   - Watches `KNOWLEDGE_DIR` and `CHATS_DIR` for file creation, modification, and deletion.
   - Extracts semantic text and metadata from 20+ file formats (including documents, spreadsheets, slides, audio tracks, video subtitles, OCR images, source code, and dialogue logs).
   - Generates normalized 1D embedding vectors via Port 8081 and atomically persists them into `~/knowledge_db.json`.

4. **Agent API Gateway (`agent.py` on Port 5000)**:
   - The central dispatch router for all external interactions.
   - Exposes `/v1/chat/completions`, `/v1/models`, `/v1/thinking`, and `/v1/local_info`.
   - Embeds incoming queries, retrieves matching knowledge and memory chunks, calculates cosine similarity, constructs augmented prompt contexts, and streams completions back to clients.
   - Hosts the air-gapped zero-dependency Web Chat UI at root `GET /`.

5. **Resilient Multi-Provider SSH Tunnel Daemon**:
   - Establishes a reverse SSH tunnel forwarding local port 5000 to a public HTTPS endpoint.
   - Automatically tests and cascades through multiple providers (`localhost.run` -> `serveo.net` -> `pinggy.io`) with self-healing reconnection.

6. **Go & Bubble Tea Telemetry Dashboard (`cmd/wola-tui`)**:
   - Modern terminal user interface built with Charm's Bubble Tea, Lip Gloss, and Bubbles.
   - Features real-time animated sparklines for token velocity, RAM usage bars, live multi-daemon log filtering, vector search inspection, and fuzzy command palette (`Ctrl+K`).

---

## 3. The Strict 1-to-1 Technical vs. Casual Command Model

### Command Pairing Rationale

A frequent point of friction in command-line systems is inconsistent naming: users must memorize whether a command is named `rm`, `remove`, `delete`, `destroy`, `purge`, `nuke`, or `selfdestruct`.

Wolarc enforces a strict rule across the entire codebase:
- **Every action has AT MOST ONE alias.**
- One **Technical Name**: Terse, compact, developer-friendly (ideal for rapid terminal typing and shell scripts).
- One **Casual Name**: Clear, intuitive, plain-English descriptive (ideal for new users, discoverability, and accessibility).
- Both names are available as standalone executables (`wola.<name>`) AND as subcommands to the unified dispatcher (`wola <name>` and `wolarc <name>`).
- Redundant 3rd, 4th, or 5th aliases (such as `nuke`, `destroy`, `gh`, `logs`, `upgrade`) have been completely removed.

### Complete Command Matrix

| Function / Purpose | Casual Name (Friendly) | Technical Name (Terse) | Dispatcher Usage | Standalone Executable |
| :--- | :--- | :--- | :--- | :--- |
| **Start Stack** | `start` | `r` | `wola start` / `wola r` | `wola.start` / `wola.r` |
| **Stop Stack** | `stop` | `k` | `wola stop` / `wola k` | `wola.stop` / `wola.k` |
| **Restart Stack** | `restart` | `rs` | `wola restart` / `wola rs` | `wola.restart` / `wola.rs` |
| **Status Telemetry** | `check` | `status` | `wola check` / `wola status` | `wola.check` / `wola.status` |
| **Live Dashboard** | `dashboard` | `d` | `wola dashboard` / `wola d` | `wola.dashboard` / `wola.d` |
| **Chat Client** | `chat` | `c` | `wola chat` / `wola c` | `wola.chat` / `wola.c` |
| **Model Inventory** | `models` | `m` | `wola models` / `wola m` | `wola.models` / `wola.m` |
| **Model Store / Hub** | `store` | `gt` | `wola store` / `wola gt` | `wola.store` / `wola.gt` |
| **Vector Knowledge** | `knowledge` | `i` | `wola knowledge` / `wola i` | `wola.knowledge` / `wola.i` |
| **Hardware Benchmark** | `bench` | `b` | `wola bench` / `wola b` | `wola.bench` / `wola.b` |
| **Log Monitor** | `log` | `l` | `wola log` / `wola l` | `wola.log` / `wola.l` |
| **Reasoning Stream** | `think` | `th` | `wola think` / `wola th` | `wola.think` / `wola.th` |
| **Remote Gateway Link** | `link` | `url` | `wola link` / `wola url` | `wola.link` / `wola.url` |
| **Offline Hotspot** | `hotspot` | `local` | `wola hotspot` / `wola local`| `wola.hotspot` / `wola.local` |
| **Tailscale Mesh VPN** | `tailscale` | `ts` | `wola tailscale` / `wola ts` | `wola.tailscale` / `wola.ts` |
| **Configuration Tuner** | `config` | `t` | `wola config` / `wola t` | `wola.config` / `wola.t` |
| **Update Checker** | `update` | `u` | `wola update` / `wola u` | `wola.update` / `wola.u` |
| **System Uninstaller** | `uninstall` | `purge` | `wola uninstall` / `wola purge`| `wola.uninstall` / `wola.purge` |
| **Interactive Wizard** | `help` | `h` | `wola help` / `wola h` | `wola.help` / `wola.h` |
| **UI Mode Switch** | `ui` | `tui` | `wola ui` / `wola tui` | `wola.ui` / `wola.tui` |

### Unified Subcommand Dispatcher (`wola` / `wolarc`)

You can invoke any capability using either syntax:
```bash
# Standalone command style
wola.start
wola.check
wola.chat "What is quantum computing?"

# Unified dispatcher style
wola start
wola check
wola chat "What is quantum computing?"

# Terse technical style
wola.r
wola.status
wola c "What is quantum computing?"
```

Every command cleanly handles `-h` and `--help` without hanging or initiating interactive prompt loops:
```bash
wola.start --help
wola.models --help
wola.bench --help
wola.chat --help
```

---

## 4. Exhaustive Command Reference Manual

### Lifecycle & Process Management

#### `wola.start` / `wola.r`
Boots the complete autonomous AI stack:
1. Validates configuration in `~/.wolarc` and verifies presence of `llama-server`.
2. Verifies chat and embedding models in `MODELS_DIR` (offers automatic quick-start download if missing).
3. Acquires Android CPU wake-lock (`termux-wake-lock`) to maintain computation with screen off.
4. Spawns embedding server on Port 8081 (`http://127.0.0.1:8081`).
5. Spawns primary chat LLM on Port 8080 (`http://127.0.0.1:8080`).
6. Starts `indexer.py` file-watching daemon.
7. Launches `agent.py` API Gateway on Port 5000 (`http://127.0.0.1:5000`).
8. Negotiates resilient reverse SSH tunnel and saves active URL in `~/.wola_url`.
9. Writes runtime state lock `~/.wola.lock`.

```bash
# Casual invocation
wola.start
# Technical invocation
wola.r
# Subcommand invocation
wola start
```

#### `wola.stop` / `wola.k`
Cleanly terminates all active services and releases system resources:
- Sends `SIGKILL` to `llama-server` processes.
- Terminates `agent.py` and `indexer.py` background daemons.
- Kills active SSH tunnel processes.
- Releases Android wake-lock (`termux-wake-unlock`).
- Removes lock and state files (`~/.wola.lock`, `~/.wola_url`, `~/.wola_endpoint`).

```bash
# Casual invocation
wola.stop
# Technical invocation
wola.k
# Subcommand invocation
wola stop
```

#### `wola.restart` / `wola.rs`
Executes an automated graceful restart sequence:
- Invokes `wola.stop` to terminate processes.
- Pauses 2 seconds for complete socket release and memory freeing.
- Invokes `wola.start` to reload stack with latest configurations.

```bash
wola.restart
wola.rs
```

#### `wola.check` / `wola.status`
Queries system health without altering runtime state:
- Displays switch state (`ON` or `OFF`).
- Reports individual process PIDs and port listening status for Chat (8080), Embed (8081), Agent (5000), Indexer, and Tunnel.
- Reports system RAM usage, active models, and recent inference speed metrics.

```bash
wola.check
wola.status
```

---

### Telemetry Dashboard & Interactive TUI

#### `wola.dashboard` / `wola.d`
Launches the full-screen terminal dashboard:
- When the Go Bubble Tea TUI is enabled, delegates directly to the high-performance compiled binary `wola-tui`.
- When classic CLI mode is active or when invoked with `--once`, displays a formatted terminal telemetry snapshot and exits immediately.

```bash
# Launch interactive dashboard
wola.dashboard
wola.d

# Single-shot telemetry snapshot (non-interactive, script-friendly)
wola.dashboard --once
wola.d --once
```

#### `wola.ui` / `wola.tui`
Controls the default user interface behavior when launching `wola` without arguments:

```bash
# Enable the Go Bubble Tea TUI as default
wola ui on
wola tui on

# Disable the TUI and default to the lightweight classic CLI suite (wola-h)
wola ui off
wola tui off

# Inspect current UI mode setting
wola ui status
wola tui status
```

---

### Conversational Clients

#### `wola.chat` / `wola.c`
Terminal chat client supporting both interactive REPL conversations and single-shot CLI queries:

```bash
# Launch interactive REPL chat session
wola.chat
wola.c

# Single-shot CLI query (streams response to stdout and exits with code 0)
wola.chat "Summarize the primary benefits of local vector embeddings"
wola.c "Explain how to configure an SSH tunnel"
```

**In-Chat Slash Commands:**
- `/exit`, `/quit`, `/q`: Save session and exit chat.
- `/clear`, `/c`: Clear active conversational context.
- `/think`: Toggle visibility of model cognitive reasoning tokens (`<think>`).
- `/temp <val>`: Dynamically set sampling temperature (e.g. `/temp 0.3`).
- `/multi`: Enter multi-line prompt mode (end input with EOF / `Ctrl+D`).

---

### Model Management & Store

#### `wola.models` / `wola.m`
Model inventory inspector and role assigner:

```bash
# List all GGUF models in ~/models with sizes and active roles
wola.models
wola.m

# Set active Chat model
wola.models chat my-model-q4_k_m.gguf
wola.m chat my-model-q4_k_m.gguf

# Set active Embedding model
wola.models embed bge-small-en-v1.5.gguf
wola.m embed bge-small-en-v1.5.gguf

# Import model from external device path
wola.models import /sdcard/Download/custom-model.gguf
```

#### `wola.store` / `wola.gt`
Interactive Model Hub for downloading verified GGUF models directly from Hugging Face:

```bash
# Launch interactive Model Store wizard
wola.store
wola.gt

# Automated non-interactive starter model download
wola.store --quick-start
wola.gt --quick-start
```

---

### Vector Knowledge Base & RAG Indexer

#### `wola.knowledge` / `wola.i`
Vector index inspector, manual document importer, and semantic search retrieval tester:

```bash
# Inspect knowledge database statistics and indexed file list
wola.knowledge status
wola.i status

# Import a document or folder into the knowledge base and index immediately
wola.knowledge add ~/Documents/report.pdf
wola.i add /path/to/notes/

# Test vector similarity search directly from CLI
wola.knowledge search "network configuration"
wola.i search "retrieval augmented generation"

# Trigger an immediate incremental scan
wola.knowledge sync
wola.i sync

# Purge vector database and re-index all documents from scratch
wola.knowledge rebuild
wola.i rebuild
```

---

### Hardware Profiling & Auto-Tuning

#### `wola.bench` / `wola.b`
Profiles CPU thread configurations to determine the optimal inference balance:

```bash
# Run comprehensive multi-thread benchmark sweep
wola.bench
wola.b

# Fast non-inference hardware recommendation check
wola.bench --check
wola.b --check

# Benchmark and automatically apply optimal settings to ~/.wolarc
wola.bench -y
wola.b --apply
```

---

### Diagnostics & Live Logging

#### `wola.log` / `wola.l`
Unified live log tailer with multiplexed or service-specific streams:

```bash
# Multiplexed live stream of all services
wola.log
wola.l

# Stream specific service log
wola.log chat      # Chat LLM engine (Port 8080)
wola.log embed     # Embedding engine (Port 8081)
wola.log agent     # API Gateway & Web UI (Port 5000)
wola.log tunnel    # Outbound SSH tunnel connection
wola.log think     # Real-time model reasoning stream
```

#### `wola.think` / `wola.th`
Controls and views model cognitive reasoning tokens (`<think>`):

```bash
# Stream internal reasoning thoughts in real time as inference generates
wola.think
wola.th

# Toggle thinking mode across the entire node
wola.think on      # Enable reasoning tokens
wola.think off     # Disable thinking tokens (saves battery, compute & latency)
wola.think status  # Check current thinking state
```

---

### Gateways & Network Connectivity

#### `wola.link` / `wola.url`
Manages public network access, tunnels, and connectivity diagnostics:

```bash
# Display active public URL, latency ping, and camera-scannable ASCII QR code
wola.link
wola.url

# Run multi-route latency & reachability benchmark across all interfaces
wola.link test

# Display comprehensive connection audit and watchdog status
wola.link status

# Switch tunnel provider directly (cloudflare, localhost.run, serveo, pinggy)
wola.link switch cloudflare
wola.link switch localhost.run

# Force tunnel reconnect & restart watchdog
wola.link reconnect

# Copy public API endpoint to clipboard
wola.link copy
wola.url copy
```

#### `wola.tailscale` / `wola.ts`
Manages Tailscale WireGuard mesh VPN peer connectivity:

```bash
# Inspect Tailscale mesh IP, connection links, and status
wola.tailscale
wola.ts
wola.link tailscale

# Display Tailscale QR code for mobile connection
wola.ts qr

# Copy Tailscale API endpoint (http://100.x.y.z:5000/v1) to clipboard
wola.ts copy

# View interactive Tailscale setup and activation guide
wola.ts guide

# Configure Tailscale Serve / Funnel
wola.ts serve
wola.ts funnel
```

#### `wola.hotspot` / `wola.local`
Manages air-gapped offline access over local network interfaces:

```bash
# Display detected local interfaces (Wi-Fi LAN, Hotspot, USB), Web UI link, and QR code
wola.hotspot
wola.local

# Display standalone ASCII QR code for phone camera scanning
wola.hotspot qr
wola.local qr

# Copy primary local URL to clipboard
wola.hotspot copy

# Output raw local URL (for shell scripts)
wola.hotspot raw

# Output OpenAI API endpoint URL
wola.hotspot api
```

---

### Configuration Parameter Tuning

#### `wola.config` / `wola.t`
View and modify runtime configurations in `~/.wolarc`:

```bash
# View all active configuration parameters
wola.config
wola.t

# Update specific parameter
wola.config TEMPERATURE 0.6
wola.t CHAT_THREADS 4
wola.t SIMILARITY_THRESHOLD 0.50
wola.t FLASH_ATTENTION auto
```

---

### Maintenance & Updates

#### `wola.update` / `wola.u`
Checks for updates from the official GitHub repository and upgrades the installation:

```bash
# Interactive check and upgrade
wola.update
wola.u

# Check for updates without applying
wola.update check

# Apply updates immediately
wola.update apply

# Inspect currently installed version commit hash
wola.update version
```

---

### Interactive Command Center

#### `wola.help` / `wola.h`
Launches the interactive terminal command center:
- Displays live node status, active URLs, and port bindings.
- Offers a numbered interactive quick-launch menu (`[1]` to `[12]`) to run any component with a single keystroke.

```bash
wola.help
wola.h
```

---

### System Cleanup & Deletion Engine

#### `wola.uninstall` / `wola.purge`
Multi-tier uninstaller and data deletion engine:

```bash
# Launch interactive hardness level selection menu
wola.uninstall
wola.purge

# Level 1: Soft Cleanup (removes binaries, configs, locks, logs; PRESERVES models & knowledge)
wola.uninstall --soft -y
wola.purge 1 -y

# Level 2: Medium Cleanup (Level 1 + deletes models and vector database; PRESERVES user notes)
wola.uninstall --medium -y
wola.purge 2 -y

# Level 3: Hard / Complete Self-Destruction (permanently erases everything including repository)
wola.uninstall --hard -y
wola.purge 3 -y
```

---

## 5. Universal RAG Pipeline & Semantic Memory

### Multi-Format Ingestion Engine

The background indexer (`indexer.py`) continuously monitors the configured knowledge folder (`KNOWLEDGE_DIR`) and conversations folder (`CHATS_DIR`). It natively extracts content from:

- **Rich Documents & E-Books**: `.pdf` (via `pdftotext` or python extractors), `.docx`, `.epub`, `.rtf`, `.odt`, `.md`, `.txt`, `.html`, `.tex`, `.org`.
- **Audio Recordings & Music**: `.mp3`, `.wav`, `.m4a`, `.ogg`, `.flac`, `.opus`, `.aac`, `.wma` (extracts ID3/metadata tags, artist, lyrics, and companion transcripts `.lrc`, `.srt`, `.txt`).
- **Video Media**: `.mp4`, `.mkv`, `.mov`, `.avi`, `.webm`, `.flv` (extracts container metadata, chapters, and companion subtitle tracks `.srt`, `.vtt`).
- **Photos & Graphics**: `.jpg`, `.jpeg`, `.png`, `.webp`, `.svg`, `.gif` (extracts EXIF tags, GPS metadata, dates, captions, and Tesseract OCR visual text when present).
- **Office Sheets & Slides**: `.pptx` (slide text), `.xlsx` (tables & cell strings), `.csv`, `.tsv`, `.ods`.
- **Source Code**: `.py`, `.js`, `.ts`, `.html`, `.css`, `.json`, `.yaml`, `.toml`, `.sql`, `.sh`, `.c`, `.cpp`, `.rs`, `.go`, `.java`.
- **Saved Conversations**: `.md`, `.chat` files from previous chats, preserved by dialogue turn so questions and answers stay semantically grouped.

### Vector Chunking & Embedding Generation

1. Files are processed through semantic sliding-window chunkers (approximately 400–600 tokens per chunk with 50-token overlap).
2. Code and dialogue logs preserve syntactic blocks and conversation turns intact.
3. Chunks are passed via HTTP POST to the local embedding server at `http://127.0.0.1:8081/embedding`.
4. The resulting vector arrays are normalized into clean 1D float vectors.

### Vector Database Architecture

The vector index is saved atomically to `~/knowledge_db.json`. Each document record contains:
```json
{
  "/absolute/path/to/document.md": {
    "filename": "document.md",
    "mtime": 1775839201.0,
    "sha256": "4a7f...",
    "file_type": "document",
    "chunks": [
      {
        "text": "Chunk text content...",
        "vector": [0.0241, -0.0185, 0.0892, ...]
      }
    ]
  }
}
```

### Cosine Similarity Scoring & Dynamic Thresholding

When a user query arrives at `/v1/chat/completions`:
1. The user query is vectorized through Port 8081.
2. Cosine similarity is computed against every chunk in `knowledge_db.json`:
   $$\text{similarity} = \frac{A \cdot B}{\|A\| \|B\|}$$
3. Chunks exceeding `SIMILARITY_THRESHOLD` (default: `0.50`) are ranked.
4. The top $K$ chunks (`TOP_K`, default: `2` to `4`) are selected for prompt augmentation.

### Memory Blending: Base Intelligence vs. Personal Context

Rather than blindly appending raw text with instructions like "Answer only using the notes", Wolarc separates context into two distinct semantic categories:

1. **`[Past Conversation Memory & User Context]`**: User habits, historical decisions, preferences, and recurring subjects from past chat sessions.
2. **`[Personal Knowledge & Memory Context]`**: Verified personal documents, codebases, media notes, and reference files.

**System Directives Instruct the LLM:**
- Prioritize its own foundational intelligence, logic, and creativity.
- Naturally weave matching memories and personal facts into the response.
- Never recite awkward disclaimers like *"According to your notes"* or *"Based on the provided text"* unless explicitly asked.

---

## 6. Offline Air-Gapped Networking, Tailscale Mesh & Remote Tunneling

Wolarc implements a multi-layer network topology designed for resilience across air-gapped field deployments, private mesh networks, and public remote access.

### Local Wi-Fi LAN Discovery

When your device is connected to a local Wi-Fi router, Wolarc detects its local IP address (e.g. `192.168.1.105`). Any device on the same Wi-Fi network can access the Web UI and API at `http://192.168.1.105:5000/`.

### Android Portable Hotspot Pairing

When travelling or completely off-grid with no Wi-Fi or cellular service:
1. Turn on Android **Portable Wi-Fi Hotspot** on your phone.
2. Connect your laptop, tablet, or another phone to your phone's hotspot Wi-Fi.
3. Run `wola.hotspot` (or `wola.local`).
4. Android assigns the host device the fixed IP `192.168.43.1`.
5. Other devices navigate directly to `http://192.168.43.1:5000/` or scan the ASCII QR code generated by `wola.hotspot qr`.
6. Full local AI chat, vector retrieval, and OpenAI API access function with **zero internet connection**.

### USB Reverse Tethering / RNDIS

Connect your phone to a PC via USB cable and enable USB Tethering. Wolarc automatically detects the RNDIS interface (`192.168.42.129:5000`), allowing ultra-low latency direct cable access without wireless interference.

### Tailscale WireGuard Mesh VPN Integration

Tailscale establishes a zero-configuration, encrypted WireGuard overlay mesh network (tailnet) connecting your phone, PC, and remote devices globally without exposing any public ports:

- **Dual-Mode OS & CLI Auto-Detection**:
  - *Android OS App Mode*: Directly inspects virtual interface IP assignments (`100.64.0.0/10` on `tun*`). If the official Android Tailscale app is active, Wolarc identifies the mesh IP and generates pairing endpoints without needing root or command-line tools.
  - *Termux / Linux CLI Mode*: When the `tailscale` binary is installed in `$PATH`, Wolarc automatically queries MagicDNS hostnames (e.g. `http://my-phone.tailnet.ts.net:5000/`), status, and enables `wola.ts serve` and `wola.ts funnel`.
- **Universal CORS Header Injection**: Third-party web clients (NextChat Web, Chatbox Web, OpenWebUI) require Cross-Origin Resource Sharing headers to reach private IP ranges. Wolarc's API Gateway automatically injects `Access-Control-Allow-Origin: *` and complete preflight options on all routes.
- **Commands**:
  - `wola.tailscale` / `wola.ts`: Displays Tailscale status, assigned IP, MagicDNS address, and pairing links.
  - `wola.ts qr`: Generates an in-terminal QR code for mobile scanning.
  - `wola.ts copy`: Copies `http://100.x.y.z:5000/v1` to system clipboard.
  - `wola.ts guide`: Displays interactive platform setup and troubleshooting instructions.

### Cloudflare Quick Tunnels

Cloudflare Quick Tunnels (`cloudflared tunnel --url http://127.0.0.1:5000`) provide high-bandwidth, zero-rate-limit public HTTPS URLs on `trycloudflare.com` without requiring an account or API token:
- If `cloudflared` is detected on the host system, Wolarc prioritizes Cloudflare as its premier remote tunnel provider.
- Full support for real-time Server-Sent Events (SSE) streaming and large token context transfers.

### Autonomous Tunnel Watchdog Daemon (`bin/wola-tunnel-daemon`)

To eliminate brittle SSH disconnects and silent drops, Wolarc runs an autonomous background watchdog daemon:
- **Health Loop**: Sends an HTTP health probe to `${url}/health` every 20 seconds.
- **Automatic Fallback Cascading**: If 2 consecutive health checks fail, the watchdog automatically terminates the stale process and negotiates a new tunnel across the provider cascade:
  $$\text{Cloudflare Quick Tunnel} \longrightarrow \text{localhost.run} \longrightarrow \text{serveo.net} \longrightarrow \text{pinggy.io}$$
- **Process Supervision**: Detaches cleanly with PID tracking (`~/.wola_tunnel_daemon.pid`), monitors parent node status via `~/.wola.lock`, and logs all connection events to `~/tunnel.log`.
- **Diagnostics**: Run `wola-tunnel-daemon status` or `wola.link status` to inspect watchdog telemetry and provider health.

### Multi-Route Reachability & Latency Benchmark (`wola.link test`)

Audits the latency and availability of all network layers simultaneously:
- Pings Localhost loopback, Wi-Fi LAN, Mobile Hotspot, Tailscale mesh, and Public tunnel.
- Tests outbound SSH/HTTPS reachability across all tunnel providers to diagnose ISP/firewall blocks.

---

## 7. Inference Engine & Deep Reasoning Controls

### Deep Thinking Mechanism (`<think>`)

Reasoning models (like DeepSeek-R1 or Qwen-2.5-Thinking) generate step-by-step cognitive deliberate traces enclosed within `<think>...</think>` tokens.

**Configuring Thinking:**
- **On (`wola.think on`)**: Models emit full cognitive deliberative traces. In the Web UI and Go TUI, these are parsed into collapsible interactive thought boxes. In the terminal, they are formatted in dimmed magenta blocks.
- **Off (`wola.think off`)**: Models are instructed with direct response directives, and the API gateway streaming filter actively strips `<think>` tokens.
  - **Saves Mobile Compute**: Reduces token generation counts by 40%–80%.
  - **Reduces Battery Drain & Heat**: Stops prolonged mobile CPU pegged usage.
  - **Minimizes Latency**: Users receive answers immediately without waiting for hundreds of hidden deliberation tokens.

### Dual Engine Isolation (Chat vs. Embed)

Running embedding generation on the same `llama-server` instance as the chat LLM causes context evictions and severe latency spikes. Wolarc solves this by running two independent `llama-server` daemons:
- **Port 8080**: Dedicated to Chat inference.
- **Port 8081**: Dedicated to Embeddings with `--embedding` flag.

---

## 8. Configuration Knobs & Environment Variables

All settings reside in `~/.wolarc`. Inspect or adjust them anytime via `wola.config KEY VALUE` or `wola.t KEY VALUE`.

| Variable | Default | Purpose & Recommendations |
| :--- | :--- | :--- |
| `CHAT_MODEL` | `qwen2.5-1.5b-...` | Primary chat LLM filename in `~/models` |
| `EMBED_MODEL` | `embeddinggemma-...`| Embedding model filename in `~/models` |
| `CHAT_PORT` | `8080` | Port for chat inference engine |
| `EMBED_PORT` | `8081` | Port for embedding inference engine |
| `AGENT_PORT` | `5000` | Port for Agent API gateway & Web UI |
| `CHAT_THREADS` | `4` | CPU cores allocated to chat (set to number of performance cores) |
| `EMBED_THREADS` | `2` | CPU cores allocated to embedding |
| `CHAT_CTX` | `4096` | Context window size in tokens (lower to 2048 on devices with < 4GB RAM) |
| `EMBED_CTX` | `2048` | Context window size for embedding chunks |
| `FLASH_ATTENTION` | `auto` | Flash Attention flag (`auto`, `on`, `off`) — drastically saves RAM |
| `MLOCK` | `false` | Pin model weights to RAM to prevent OS swap thrashing (`true`/`false`) |
| `CHAT_PARALLEL` | `2` | Number of simultaneous execution slots |
| `TEMPERATURE` | `0.6` | Generation temperature (`0.6` recommended for natural reasoning) |
| `SIMILARITY_THRESHOLD` | `0.50` | Minimum cosine similarity required to inject a chunk |
| `TOP_K` | `2` | Maximum document chunks injected into prompt context |
| `ENABLE_THINKING` | `true` | Enable or disable `<think>` reasoning traces |
| `ENABLE_TUI` | `true` | Launch Go Bubble Tea TUI on `wola` (`true`) or classic CLI (`false`) |
| `TUNNEL_PROVIDER` | `auto` | Tunnel provider (`auto`, `localhost.run`, `serveo`, `pinggy`) |
| `SYSTEM_PROMPT` | Custom | Foundational system persona instructions |

---

## 9. Third-Party Client App Integration Guide

Wolarc's API Gateway on Port 5000 is 100% OpenAI API compatible.

### General Settings
- **Base URL**: `https://<your-subdomain>.lhr.life/v1` (Remote) or `http://192.168.43.1:5000/v1` (Hotspot)
- **API Key**: `sk-test` (or any string)
- **Model**: Leave auto-detected or enter your `CHAT_MODEL` name

### Connecting Specific Apps

#### Chatbox (Mobile & Desktop)
1. Open Chatbox Settings -> **Model Provider** -> Select **OpenAI API**.
2. Set **API Host**: `https://<your-subdomain>.lhr.life` (or your local IP: `http://192.168.1.X:5000`).
3. Set **API Key**: `sk-test`.
4. Click **Check** to verify connectivity.

#### Obsidian (Smart Connections & Copilot)
1. In Obsidian Settings -> Community Plugins -> **Smart Connections** or **Copilot**.
2. Select **Custom OpenAI Compatible** provider.
3. Set **Chat Endpoint**: `https://<your-subdomain>.lhr.life/v1/chat/completions`.
4. Set **API Key**: `sk-test`.

#### Python OpenAI SDK
```python
from openai import OpenAI

client = OpenAI(
    base_url="https://<your-subdomain>.lhr.life/v1",  # or http://127.0.0.1:5000/v1
    api_key="sk-test"
)

response = client.chat.completions.create(
    model="default",
    messages=[{"role": "user", "content": "Hello Wola!"}],
    stream=True
)

for chunk in response:
    content = chunk.choices[0].delta.content or ""
    print(content, end="", flush=True)
print()
```

---

## 10. Daemon Lifecycle & Operating System Integration

### Android Wake-Lock Integration
Android OS aggressively suspends background CPU cycles when the device screen locks. Wolarc integrates with the Termux API:
- On startup (`wola.start`), Wolarc executes `termux-wake-lock`, requesting a high-performance wakelock from Android OS.
- On clean shutdown (`wola.stop`), Wolarc executes `termux-wake-unlock`, allowing Android to return to normal power management.

### Detached Process Hierarchy & Signal Traps
To ensure services stay online when the user closes their terminal app:
- Background processes are launched with `setsid -f` or background job detachment with stdout/stderr redirected to log files.
- Process PIDs are mapped and monitored.
- Signal traps catch `SIGINT` and `SIGTERM` to perform atomic state teardowns without leaving orphan socket locks.

---

## 11. Future Milestones & System Roadmap

Wolarc's evolution is centered on creating a seamless, distributed personal AI ecosystem while uncompromisingly maintaining local data sovereignty. For full architectural diagrams and phase details, see [**ROADMAP.md**](ROADMAP.md).

### Milestone 1: Reliable & Effective Client-Host Connection Method
Edge mobile nodes operate across dynamic networks with shifting IP addresses and Carrier-Grade NAT (CGNAT). The future connection fabric establishes an unbreakable link:
1. **Zero-Configuration Local Discovery (mDNS / Zeroconf)**:
   - Host announces its presence via `_wolarc._tcp.local` on port 5000.
   - Client applications instantly discover available Wolarc nodes on the LAN or mobile hotspot without manual IP configuration.
2. **Resilient Multi-Transport Mesh with Auto-Failover**:
   - **Tier 1 (LAN)**: Sub-2ms low-latency connection for local Wi-Fi and offline Android hotspots.
   - **Tier 2 (Tailscale WireGuard Mesh)**: Direct peer-to-peer encrypted WireGuard tunnels across NATs via STUN/DERP.
   - **Tier 3 (Managed Cloudflare Named Tunnels)**: Authenticated persistent tunnels (`cloudflared`) with custom static hostnames.
   - **Tier 4 (WebRTC DataChannels)**: Peer-to-peer data channels bypassing CGNAT for direct client-host streaming.
3. **Resumable Token Streaming & Session Continuity**:
   - Resumable SSE streams via `Last-Event-ID` buffers, ensuring mobile network transitions (e.g. Wi-Fi dropping to 5G) do not terminate ongoing LLM token generation or drop context.
   - Stateful ping/pong heartbeats detecting disconnects within 3 seconds.
4. **Cryptographic Out-of-Band Pairing Protocol**:
   - Ephemeral Ed25519 pairing key exchange rendered as an ASCII QR code or numeric PIN (`wola.pair`).
   - Scoped authorization tokens (`chat`, `knowledge`, `admin`) stored in `~/.wola/run/authorized_clients.json`.

### Milestone 2: Native Cross-Platform Flutter Companion App
An official, high-performance companion app engineered in Flutter (Dart) targeting Android (APK), iOS, Linux, Windows, and macOS:
1. **Dynamic Connection Hub**:
   - Instant camera QR code scanner for one-tap pairing with Wolarc host nodes.
   - Auto-discovers local nodes via mDNS.
   - Multi-node manager: switch seamlessly between phone nodes, local workstations, and remote home servers.
2. **Deep Thinking & Streaming Conversational UI**:
   - Smooth SSE streaming tokens with real-time velocity metrics (`tok/s`).
   - Expandable, styled `<think>` cognitive reasoning blocks with duration counters and step logs.
   - Rich Markdown, syntax highlighting across 60+ programming languages, and LaTeX math.
3. **Universal Document & Media Uploader (RAG)**:
   - Native OS share sheet and document picker to upload PDFs, Word docs, notes, audio, and images directly into `~/Documents/wola/knowledge/`.
   - Real-time indexing progress and vector database inspection.
   - Response citation inspector: tap in-chat sources to view original document chunks and similarity match scores.
4. **Hardware Telemetry & Node Remote Control**:
   - Live gauges for CPU load, RAM allocation, battery levels, and thermal conditions.
   - Remote daemon management (Start, Stop, Restart) and live multiplexed log streams.
   - Remote model store browser with one-tap Hugging Face GGUF downloads and runtime parameter sliders.

### Phased Delivery Schedule
- **Phase 1**: Connection protocol specification, mDNS broadcast daemon, and Ed25519 handshake.
- **Phase 2**: Resumable SSE stream buffer, named Cloudflare tunnels, and `wola.pair` / `wola.peers` CLI tools.
- **Phase 3**: Flutter project initialization, Cyber-Wolf Dark design system, and QR/mDNS pairing UI.
- **Phase 4**: Flutter chat interface with SSE streaming and collapsible `<think>` blocks.
- **Phase 5**: Knowledge base uploader, vector explorer, and telemetry dashboard in Flutter.
- **Phase 6**: Remote model store, parameter tuner, and cross-platform releases (Android APK, Linux, Windows, macOS).


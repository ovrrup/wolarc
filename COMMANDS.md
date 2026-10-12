# Wolarc Node Suite — Complete Command & Architecture Reference

> **Wolarc** is an autonomous, high-performance, local RAG (Retrieval-Augmented Generation) AI node designed for Termux (Android) and Linux. It exposes an OpenAI-compatible API, indexes local knowledge notes in real time, and establishes resilient public tunnels for remote access.

---

## Command Quick Index (Casual / Technical Pairs)

> **Strict 1-to-1 Pairing Rule**: Every command in Wolarc has exactly **ONE Casual Name** (friendly, descriptive) and **ONE Technical Name** (terse, developer-friendly). Redundant legacy aliases have been completely removed.

| Feature / Action | Casual Name | Technical Name | Subcommands | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Unified TUI / Suite** | `wola` | `wola-tui` | `wola` / `wola-tui` | Unified Go & Bubble Tea TUI (`dashboard`, `chat`, `models`, `knowledge`, `logs`) |
| **UI Mode Switch** | `wola.ui` | `wola.tui` | `wola ui` / `wola tui` | Toggle default interface mode (`on` / `off` / `status`) between Go TUI and classic CLI |
| **Start Services** | `wola.start` | `wola.r` | `wola start` / `wola r` | Start all AI node services (chat, embed, API, tunnel, wake-lock) |
| **Stop Services** | `wola.stop` | `wola.k` | `wola stop` / `wola k` | Stop all AI node services cleanly and release wake-locks |
| **Restart Stack** | `wola.restart` | `wola.rs` | `wola restart` / `wola rs` | Graceful restart cycle (shutdown -> cooldown -> boot) |
| **Service Status** | `wola.check` | `wola.status` | `wola check` / `wola status` | Display active process PIDs, listening ports, and switch status |
| **Telemetry Dashboard**| `wola.dashboard`| `wola.d` | `wola dashboard` / `wola d` | Full-screen telemetry dashboard (or `--once` for instant snapshot) |
| **Interactive Chat** | `wola.chat` | `wola.c` | `wola chat` / `wola c` | Interactive CLI chat or one-shot query (`wola.chat "query"`) |
| **Model Inventory** | `wola.models` | `wola.m` | `wola models` / `wola m` | List models in `~/models` or switch active Chat/Embed roles |
| **Model Store / Hub** | `wola.store` | `wola.gt` | `wola store` / `wola gt` | Model Store & local GGUF file importer (HuggingFace / storage) |
| **Vector Knowledge Base**| `wola.knowledge`| `wola.i` | `wola knowledge` / `wola i` | Vector indexer manager (`status`, `add <path>`, `search <query>`) |
| **Hardware Benchmark** | `wola.bench` | `wola.b` | `wola bench` / `wola b` | Benchmark tokens/sec across CPU threads and auto-tune config |
| **Log Monitoring** | `wola.log` | `wola.l` | `wola log` / `wola l` | Live multiplexed log viewer for chat, embed, agent, and tunnels |
| **Thinking Reasoning** | `wola.think` | `wola.th` | `wola think` / `wola th` | Toggle reasoning tokens (`on` / `off` / `status`) or follow stream |
| **Remote Gateway Link**| `wola.link` | `wola.url` | `wola link` / `wola url` | Show public tunnel URL, QR code, ping latency & copy endpoint |
| **Local Hotspot Gateway**| `wola.hotspot`| `wola.local` | `wola hotspot` / `wola local`| Offline local gateway, zero-dependency Web Chat UI, and QR pairing |
| **Tailscale Mesh VPN** | `wola.tailscale`| `wola.ts` | `wola tailscale` / `wola ts` | Encrypted WireGuard mesh peer connectivity, MagicDNS, Serve & Funnel |
| **Parameter Tuner** | `wola.config` | `wola.t` | `wola config` / `wola t` | View and edit runtime inference hyperparameters in `~/.wolarc` |
| **Software Update** | `wola.update` | `wola.u` | `wola update` / `wola u` | Check for updates, pull releases, and upgrade binaries |
| **Command Center Wizard**| `wola.help` | `wola.h` | `wola help` / `wola h` | Interactive menu wizard with status overview and quick launchers |
| **System Uninstaller** | `wola.uninstall`| `wola.purge` | `wola uninstall` / `wola purge`| Multi-level file deletion, cleanup, and complete self-destruction |

---

## 1. System Lifecycle & Process Management

### `wola.r`
Boots the complete autonomous AI stack:
1. Reads configuration from `~/.wolarc`.
2. Verifies the existence of `llama-server` and designated chat/embedding GGUF models.
3. Automatically computes CPU thread allocations and applies hardware optimizations (Flash Attention, memory locking).
4. Acquires an Android CPU wake-lock (`termux-wake-lock`) to keep background compute active with screen off.
5. Launches the embedding server on port `8081` (`http://127.0.0.1:8081`).
6. Launches the primary chat server on port `8080` (`http://127.0.0.1:8080`).
7. Starts `indexer.py` background file watcher daemon.
8. Launches `agent.py` Flask/Gunicorn API Gateway on port `5000` (`http://127.0.0.1:5000`).
9. Negotiates an outbound resilient SSH tunnel through `localhost.run`, `serveo.net`, or `pinggy.io`.
10. Writes runtime lock `~/.wola.lock` and saves public gateway URL in `~/.wola_url`.

```bash
# Start all services
wola.r
# or
wolarc start
```

### `wola.k`
Cleanly terminates all running services and releases system resources:
- Sends `SIGKILL` to `llama-server` processes.
- Terminates `agent.py` and `indexer.py` daemons.
- Kills active SSH tunnel forwarders.
- Releases Android wake-lock (`termux-wake-unlock`).
- Removes `~/.wola.lock`, `~/.wola_url`, `~/.wola_endpoint`, and `~/.wola_provider`.

```bash
# Stop all services
wola.k
# or
wolarc stop
```

### `wola.rs`
Performs an automated graceful restart cycle:
- Shuts down all processes using `wola.k`.
- Waits 2 seconds for socket teardown and memory cleanup.
- Calls `wola.r` to boot the full stack with updated configurations.

```bash
# Restart stack
wola.rs
```

### `wolarc status`
Queries the live system without altering state:
- Switch status (`ON` if lock exists, `OFF` if idle).
- PIDs of Chat Server, Embedding Server, Agent API, and Indexer Daemon.
- Port availability and listening status.

---

## 2. Telemetry Dashboard & Modern TUI

### `wola.d` / `wola-tui`
Launches the full-screen terminal user interface written in Go and Bubble Tea.

#### Views & Navigation:
- **`[1]` Dashboard View**:
  - Live animated ASCII sparklines for inference speed (tokens/second) and system RAM.
  - Gauge bars with color gradients for RAM and Context Window utilization.
  - Radar pulse indicator for gateway tunnel connectivity.
  - Active chat and embedding model details.
  - Vector database metrics (total chunks, source documents).
- **`[2]` Models View**:
  - Browser of local GGUF models stored in `~/models`.
  - File size, quantization format, and assigned roles.
  - One-key role assignment for Chat and Embedding models.
- **`[3]` Knowledge View**:
  - Live document catalog with chunk counts and index timestamps.
  - Trigger manual re-index scans with `[i]`.
- **`[4]` Logs View**:
  - Live streaming log monitor for Chat, Embed, Agent, Tunnel, and Thinking logs.
  - Live substring filter drawer: press `/` to filter logs by keywords (e.g. `POST`, `error`, `eval`).
  - Auto-scroll toggle with `[s]`.

#### Global Keyboard Shortcuts:
| Key | Action |
| :--- | :--- |
| `1` / `2` / `3` / `4` | Switch directly between Dashboard, Models, Knowledge, and Logs |
| `Tab` / `Shift+Tab` | Cycle sequentially through views |
| `Ctrl+K` | Open Fuzzy Command Palette modal |
| `?` | Toggle Keybindings & Help cheat sheet modal |
| `t` | Toggle in-TUI Parameter Tuner drawer (adjust Temperature, Threads, etc.) |
| `r` | Start / Restart Wolarc services |
| `k` | Stop Wolarc services |
| `q` / `Ctrl+C` | Exit TUI |

### `wola tui`
Toggles default interface behavior when running `wola` without arguments:
- `wola tui on`: Enables Go Bubble Tea TUI as the primary interface (`ENABLE_TUI="true"` in `~/.wolarc`).
- `wola tui off`: Disables TUI and delegates `wola` to the classic CLI command center (`wola-h`).
- `wola tui status`: Displays current toggle status.

---

---

## 3. Remote Gateway, Tailscale Mesh & Connectivity Suite

Wolarc features a multi-tiered connectivity architecture supporting local air-gapped hotspots, encrypted peer-to-peer WireGuard mesh VPNs via Tailscale, and persistent public HTTPS tunnels managed by an autonomous background watchdog daemon.

### `wola.link` / `wola.url`
Manages public network access, tunnels, and connectivity diagnostics.

> **Command Pair**: `wola.link` (Casual) <-> `wola.url` (Technical). Subcommands: `wola link` / `wola url`.

```bash
# Display active public URL, ping latency, and terminal QR code
wola.link
# or
wola.url

# Run multi-route latency & reachability benchmark across all interfaces
wola.link test

# Display comprehensive connection audit and watchdog status
wola.link status

# Switch tunnel provider directly (cloudflare, localhost.run, serveo, pinggy)
wola.link switch cloudflare
wola.link switch localhost.run

# Force immediate tunnel renegotiation & watchdog restart
wola.link reconnect

# Render terminal QR code for phone camera scanning
wola.link qr

# Copy active public /v1 API endpoint to clipboard
wola.link copy

# Output raw base URL or /v1 endpoint (for shell scripts)
wola.link raw
wola.link api
```

#### Multi-Route Benchmark & Audit (`wola.link test`):
Performs an automated latency ping and port sweep across every available layer:
- **Localhost Loopback**: Tests local `127.0.0.1:5000/health` response latency.
- **Local Wi-Fi LAN**: Benchmarks local Wi-Fi interface latency.
- **Mobile Hotspot**: Tests reachability when serving as an access point.
- **Tailscale Mesh**: Measures WireGuard peer latency over Tailscale.
- **Public Tunnel**: Pings the active public HTTPS endpoint through Cloudflare or SSH.
- **Outbound Provider Sweep**: Probes outbound ports 22/443 reachability to determine which tunnel providers are accessible on your current ISP or cellular network.

---

### `wola.tailscale` / `wola.ts`
Tailscale WireGuard Mesh VPN integration for secure, direct peer-to-peer access from any device worldwide without public tunnels or port forwarding.

> **Command Pair**: `wola.tailscale` (Casual) <-> `wola.ts` (Technical). Subcommands: `wola tailscale` / `wola ts` / `wola.link tailscale`.

```bash
# Inspect Tailscale mesh IP, connection links, and status
wola.tailscale
# or
wola.ts
# or
wola.link tailscale

# Display Tailscale QR code for mobile connection
wola.ts qr

# Copy Tailscale API endpoint (http://100.x.y.z:5000/v1) to clipboard
wola.ts copy

# View interactive Tailscale setup and activation guide
wola.ts guide

# Configure Tailscale Serve (exposes Wola locally on your tailnet)
wola.ts serve

# Configure Tailscale Funnel (public HTTPS over Tailscale infrastructure)
wola.ts funnel
```

#### Tailscale Dual-Mode Operation:
- **Android OS VPN App Support**: Detects Tailscale running via the official Android VPN client by inspecting interface IP assignments (`100.64.0.0/10` on `tun*`), requiring zero root privileges or command-line setup.
- **Termux & Linux CLI Support**: When the `tailscale` CLI is present in `$PATH`, Wolarc automatically queries MagicDNS hostnames (e.g., `http://my-phone.tailnet-xyz.ts.net:5000/`), status states, and supports direct `serve` and `funnel` orchestration.
- **Universal CORS Header Support**: Automatically injects permissive CORS headers (`Access-Control-Allow-Origin: *`) on all responses, allowing browser-based AI clients (NextChat Web, Chatbox Web) hosted anywhere to connect directly to your Tailscale mesh IP.

---

### Autonomous Tunnel Watchdog Daemon (`bin/wola-tunnel-daemon`)
A resilient, self-healing background supervisor that ensures remote connectivity never drops:

- **Automated Lifecycle**: Starts automatically when Wolarc boots (`wola.start` / `wola.r`) and terminates cleanly when stopped (`wola.stop` / `wola.k`).
- **Autonomous Health Checks**: Pings `${url}/health` every 20 seconds.
- **Automatic Multi-Provider Cascading**: If 2 consecutive health checks fail, the watchdog automatically cascades to the next available provider:
  1. **Cloudflare Quick Tunnels** (`trycloudflare.com`): Zero rate limits, high bandwidth, and robust HTTPS via `cloudflared`.
  2. **localhost.run**: Low-latency SSH reverse tunnel.
  3. **serveo.net**: Resilient secondary SSH tunnel.
  4. **pinggy.io**: Alternative fallback tunnel.
- **Persistent Process Management**: Tracks background PIDs in `~/.wola_tunnel_daemon.pid` and logs all health events to `~/tunnel.log`.

```bash
# Inspect watchdog daemon state and provider health
wola-tunnel-daemon status

# Manually trigger a watchdog reconnect cascade
wola-tunnel-daemon reconnect
```

---

### `wola.hotspot` / `wola.local`
Provides air-gapped local network connectivity and pairing for offline environments:

> **Command Pair**: `wola.hotspot` (Casual) <-> `wola.local` (Technical). Subcommands: `wola hotspot` / `wola local` / `wola.link local`.

```bash
# Display local IP interfaces, Web UI link, and terminal QR code
wola.hotspot
# or
wola.local

# Generate standalone QR code in terminal for mobile scanning
wola.hotspot qr

# Copy primary local URL to clipboard
wola.hotspot copy

# Output raw local URL (for scripting)
wola.hotspot raw

# Print /v1 API endpoint
wola.hotspot api
```

#### Multi-Interface Auto-Detection:
- **Local Wi-Fi LAN**: Detects assigned IP on active wireless networks (e.g., `http://192.168.1.105:5000/`).
- **Android AP Hotspot**: Dedicated offline pairing address for when your Android device hosts its own Wi-Fi hotspot (`http://192.168.43.1:5000/`).
- **USB Reverse Tethering / RNDIS**: Direct cable connection address (`http://192.168.42.129:5000/`).
- **Network Discovery Endpoint**: Exposes `/v1/local_info` returning JSON metadata of all detected interfaces, Tailscale mesh state, and active tunnel URLs.

### Standalone Zero-Dependency Web Chat UI
Accessible directly via any browser connected to your local network, hotspot, Tailscale mesh, or public tunnel at `http://<ip>:5000/`:
- **100% Air-Gapped**: Zero external CDN calls, remote scripts, or Google Fonts. Functions completely offline without internet access.
- **Embedded Web Client**: Fully responsive dark-mode UI with markdown formatting, syntax highlighting, and dynamic copy-to-clipboard buttons.
- **Real-Time SSE Streaming**: Token-by-token live generation with elapsed time and speed metrics.
- **Deep Thinking Toggle**: Automatically parses and collapses `<think>...</think>` internal cognitive reasoning traces into interactive collapsible process boxes.
- **Quick Connection Modals**: Direct copyable snippets for curl, Python SDK, Obsidian, Chatbox, NextChat, and Tailscale Mesh links.

---

## 4. Interactive Terminal Chat Client

### `wola.chat`
A terminal chat client that interacts directly with the local AI node.

```bash
# Launch interactive chat
wola.chat
# or
wola.c
```

#### Capabilities:
- **Streaming Tokens**: Real-time rendering of generated response text.
- **RAG Note Citations**: Automatically renders cited notes and documents retrieved by vector similarity, including source file names and similarity scores.
- **In-Chat Commands**:
  - `/multi`: Enter multi-line message mode (finish with `EOF`).
  - `/reset`: Clear conversational history context.
  - `/think`: Toggle visibility of internal reasoning tokens (`<think>...</think>`).
  - `/temp <val>`: Dynamically adjust sampling temperature (e.g. `/temp 0.3`).
  - `/sys <prompt>`: Update the system instructions for the active session.
  - `/exit` or `/quit`: Exit chat client.

---

## 5. Model Thinking & Reasoning Engine

### `wola think` / `wola.th`
Controls and monitors the cognitive reasoning mechanism (`<think>` tokens) of reasoning models (such as DeepSeek-R1, Qwen-2.5, etc.).

```bash
# Toggle thinking mode on or off across the node
wola think on          # Enable internal reasoning tokens
wola think off         # Disable thinking to save compute, RAM, and time
wola think status      # Inspect current thinking configuration

# Direct alias commands
wola.th on
wola.th off
wola.th status

# Follow real-time reasoning stream live
wola.th
```

#### Why Disable Thinking?
- **Compute & Battery Savings**: Suppresses internal `<think>` scratchpads, dramatically reducing token counts, device heating, and power drain on mobile processors.
- **Immediate Response Speed**: Answers begin streaming instantly without waiting through dozens or hundreds of hidden cognitive deliberation tokens.
- **Config Persistence**: Toggling thinking updates `ENABLE_THINKING` in `~/.wolarc` and syncs dynamically with the active agent server via `/v1/thinking`.
- **Live Stream Viewer**: Running `wola.th` without arguments attaches to `~/.wola_thinking.log` to stream the model's inner monologue live as it thinks.

---

## 6. Model Store & Local GGUF Importer

### `wola.store` / `wola.gt`
Interactive Model Hub for downloading and managing GGUF models.

```bash
# Launch interactive Model Hub
wola.store
# or
wola.gt
```

#### Options:
- **Curated Model Tiers**:
  - Tier 1: Ultra-compact (1B–1.5B parameters) for low RAM devices (< 4GB).
  - Tier 2: Balanced (3B parameters) for standard mobile devices.
  - Tier 3: High Capacity (7B–8B parameters) for high-spec devices.
  - Tier 4: Embedding models (bge-micro, bge-small, nomic-embed).
- **Import Own GGUF File**:
  - Allows selecting custom `.gguf` files from device storage (e.g. `/sdcard/Download`).
  - Prompts whether to **Move** or **Copy** the model into `~/models`.
  - Automatically updates `CHAT_MODEL` or `EMBED_MODEL` in `~/.wolarc`.

### `wola.m`
Command-line model utility.

```bash
# List all models in ~/models with sizes and active roles
wola.m

# Quick CLI import from custom path
wola.m import /path/to/my-custom-model.gguf

# Delete a model interactively (shows numbered list with sizes & active tags)
wola.m rm

# Delete a model directly by filename
wola.m rm qwen2.5-1.5b-instruct-q4_k_m.gguf
# or
wola.m delete my-model.gguf -y
```

---

## 7. Vector Knowledge Base & RAG Indexer

### `wola.i`
CLI tool to manage and test the vector knowledge database.

```bash
# View database health, total indexed chunks, and file count
wola.i status

# Index a single document or an entire directory immediately
wola.i add ~/storage/shared/Documents/wola/knowledge/my_notes.md
wola.i add /path/to/folder/

# Delete a knowledge document and atomically purge its vectors from DB
wola.i rm my_notes.md
# or interactively select:
wola.i rm

# Delete a specific saved chat history and strip its dialogue vectors
wola.i rm-chat session_20261012_180000

# Clear all saved chat sessions and remove all conversation vectors
wola.i clear-chats

# Test vector retrieval directly with a search query
wola.i search "How do I configure network settings?"
```

#### Automatic Universal Indexing Daemon (`indexer.py`):
Runs in the background watching both `KNOWLEDGE_DIR` (default: `~/storage/shared/Documents/wola/knowledge` or `agent_knowledge`) and `CHATS_DIR` (default: `~/storage/shared/Documents/wola/chats` or `agent_chats`). When files or dialogues are created, edited, or removed, the indexer automatically extracts semantic content, chunks it, generates vector embeddings via port `8081`, and atomically saves them to `~/.wola/data/knowledge_db.json`.

**Supported File Formats (Zero Limitations):**
- **Documents & E-Books**: `.pdf`, `.docx`, `.epub`, `.rtf`, `.odt`, `.md`, `.txt`, `.html`, `.tex`, `.org`
- **Audio Recordings & Music**: `.mp3`, `.wav`, `.m4a`, `.ogg`, `.flac`, `.opus`, `.aac`, `.wma` (extracts ID3/RIFF/Vorbis metadata, duration, artist, lyrics, and companion transcripts `.lrc`/`.srt`/`.txt`)
- **Video Media**: `.mp4`, `.mkv`, `.mov`, `.avi`, `.webm`, `.flv`, `.wmv` (extracts container metadata, resolution, chapters, embedded/companion subtitles `.srt`/`.vtt`)
- **Photos & Graphics**: `.jpg`, `.jpeg`, `.png`, `.webp`, `.svg`, `.gif`, `.bmp` (extracts EXIF camera tags, date taken, GPS coordinates, dimensions, captions, and Tesseract OCR visual text)
- **Office Sheets & Slides**: `.pptx` (slide text), `.xlsx` (tables & cell strings), `.csv`, `.tsv`, `.ods`
- **Source Code & Configs**: `.py`, `.js`, `.ts`, `.html`, `.css`, `.json`, `.yaml`, `.toml`, `.sql`, `.sh`, `.c`, `.cpp`, `.rs`, `.go`, `.java`, etc.
- **Chat Conversations & Memories**: `.md`, `.chat` files in `CHATS_DIR` (chunked by dialogue turns so questions and answers stay together)
- **Universal Fallback**: Any text or binary data file is automatically analyzed and indexed.

#### Conversation Memory & Self-Learning:
Every chat session conducted through the Web UI, Go TUI, CLI (`wola.chat`), or API (`/v1/chat/completions`) is automatically written to `CHATS_DIR` as a markdown log (`chat_<session_id>.md`). Dialogue turns are asynchronously embedded into the vector store. When the user interacts in future sessions, relevant historical exchanges are retrieved as `[Past Conversation Memory & User Context]`, allowing the model to naturally recall past decisions, project details, and user habits.

#### Personalized Memory & Base Intelligence Blending:
The local model prioritizes its own foundational intelligence, logic, and creativity. Retrieved memories are separated into distinct contexts:
1. `[Past Conversation Memory & User Context]`: User habits, history, instructions, and preferences from past chats.
2. `[Personal Knowledge & Memory Context]`: Knowledge base notes, media, code, and documentation.

The model seamlessly blends both contexts with its core intelligence to deliver tailored, deeply personalized responses without robotic recitation or awkward "according to the notes" disclaimers.

---

## 8. Hardware Benchmarking & Optimization

### `wola.bench`
Hardware profiling tool for mobile and embedded CPUs.

```bash
# Run benchmark sweep across thread configurations
wola.bench
```

- Measures Prompt Processing (PP) and Token Generation (TG) speeds across 1 to 8 CPU threads.
- Identifies the thermal and performance sweet spot for your hardware architecture.
- Offers one-click interactive application of optimal `CHAT_THREADS` and `CHAT_CTX` values directly to `~/.wolarc`.

---

## 9. Parameter Tuner & Configuration

### `wola.t`
View and modify runtime settings in `~/.wolarc`.

```bash
# View current configuration parameters
wola.t

# Update a specific configuration parameter
wola.t TEMPERATURE 0.7
wola.t CHAT_THREADS 4
wola.t FLASH_ATTENTION auto
```

#### Key Configuration Knobs:
| Variable | Default | Description |
| :--- | :--- | :--- |
| `CHAT_MODEL` | — | Primary LLM GGUF filename located in `~/models` |
| `EMBED_MODEL` | — | Embedding GGUF filename located in `~/models` |
| `CHAT_PORT` | `8080` | Port for chat inference engine |
| `EMBED_PORT` | `8081` | Port for embedding inference engine |
| `AGENT_PORT` | `5000` | Port for OpenAI-compatible gateway |
| `CHAT_THREADS` | `4` | CPU cores allocated to chat generation |
| `EMBED_THREADS` | `2` | CPU cores allocated to embedding generation |
| `CHAT_CTX` | `4096` | Context window size (tokens) for chat |
| `EMBED_CTX` | `2048` | Context window size (tokens) for embedding |
| `FLASH_ATTENTION` | `auto` | Flash Attention flag (`auto`, `on`, `off`) |
| `MLOCK` | `false` | Pin model memory to prevent OS swap thrashing (`true`/`false`) |
| `TEMPERATURE` | `0.7` | Sampling temperature |
| `TOP_P` | `0.9` | Top-P nucleus sampling threshold |
| `TOP_K` | `40` | Top-K sampling pool |
| `RAG_TOP_K` | `4` | Number of document chunks retrieved per query |
| `RAG_SIM_THRESHOLD` | `0.35` | Minimum cosine similarity score for retrieval inclusion |

---

## 10. Live Log Monitoring & Diagnostics

### `wola.log`
Multiplexed or single-stream log monitor.

```bash
# Unified log monitor (streams all services simultaneously)
wola.log

# Stream chat server log only
wola.log chat

# Stream embedding server log only
wola.log embed

# Stream agent API gateway log only
wola.log agent

# Stream tunnel connection log only
wola.log tunnel

# Stream thinking reasoning stream log only
wola.log think

# Clear and truncate all log files to 0 bytes
wola.log clear
```

---

## 11. Software Updates & Maintenance

### `wola.u`
Checks for updates from GitHub and upgrades the local installation.

```bash
# Check and apply updates
wola.u
```

- Compares local Git commit hash (`~/.wola/run/wola_version`) against remote repository HEAD.
- Automatically preserves custom user paths and settings in `~/.wolarc`.
- Refreshes command executables and symlinks in `$PREFIX/bin` or `~/.local/bin`.
- Reconciles Python dependencies.

---

## 12. Interactive Command Center Wizard

### `wola-h`
Interactive command-line control center.

```bash
# Launch command center
wola-h
# or
wola.h
```

Displays live node status, active public URL, and an interactive menu of numbered shortcuts (`[1]` to `[11]`) to launch any component of the suite with a single keystroke.

---

## 13. System File Deletion & Self-Destruct Engine

### `wola.uninstall` / `wola.purge`
The system deletion and cleanup engine.

> **Command Pair**: `wola.uninstall` (Casual) <-> `wola.purge` (Technical). Subcommands: `wola uninstall` / `wola purge`.

This command stops all running processes, removes lock files, and erases files and directories associated with Wolarc. It provides **3 distinct hardness levels** alongside **granular targeted purge options**.

```bash
# Syntax
wola.uninstall [OPTIONS] [LEVEL]
# or
wola.purge [OPTIONS] [LEVEL]
```

### Granular Purge Flags

Target specific parts of the system without affecting the rest:

```bash
# Truncate / delete all logs
wola.purge --logs

# Remove caches, telemetry, and python bytecode
wola.purge --cache

# Delete saved chat histories and dialogue vectors
wola.purge --chats

# Delete vector database (knowledge_db.json)
wola.purge --db

# Delete all GGUF models in models directory
wola.purge --models
```

### Hardness Levels

```
Level 1 (Soft)       ──> Deletes binaries, configs, locks, daemons and logs.
                         PRESERVES: Models, Vector DB, Knowledge notes.

Level 2 (Medium)     ──> Level 1 + Deletes Models & Vector DB.
                         PRESERVES: User Knowledge notes.

Level 3 (Hard)       ──> Complete Self-Destruction.
                         ERASES: Binaries, configs, logs, models,
                         Vector DB, Knowledge notes, and repo.
```

#### Level 1: Soft (Default Cleanup)
- **Flag**: `1`, `--soft`, `--level 1`, `--level=1`
- **Actions Performed**:
  1. Stops all running processes (`llama-server`, `agent.py`, `indexer.py`, SSH tunnels, `wola-tui`).
  2. Releases Android wake-lock (`termux-wake-unlock`).
  3. Deletes runtime state and lock files (`~/.wola/run`, `.wola.lock`, `.wola_url`, `.wola_metrics.json`, etc.).
  4. Deletes log files (`~/.wola/logs`, `chat.log`, `embed.log`, `agent.log`, `tunnel.log`, `indexer.log`).
  5. Removes deployed daemon scripts (`~/.wola/lib`) and Python cache (`__pycache__`).
  6. Removes configuration file `~/.wolarc` and its backups.
  7. Deletes all installed executables and symlinks in `$TARGET_BIN` (`wolarc`, `wola.r`, `wola.k`, `wola.rs`, `wola.d`, `wola.url`, `wola.chat`, `wola.bench`, `wola.log`, `wola.th`, `wola.gt`, `wola.gh`, `wola.t`, `wola.m`, `wola.i`, `wola.u`, `wola-h`, `wola-tui`).
- **PRESERVED**:
  - `~/models` (downloaded GGUF models are kept intact)
  - `~/.wola/data/knowledge_db.json` (vector database is kept intact)
  - Knowledge notes directory (user documents are kept intact)

```bash
# Execute Level 1 Cleanup
wola.purge --soft

# Non-interactive / unattended
wola.purge --soft -y
```

---

#### Level 2: Medium (System + Models Cleanup)
- **Flag**: `2`, `--medium`, `--level 2`, `--level=2`
- **Actions Performed**:
  - Performs all actions from **Level 1**.
  - **Erases downloaded models**: Permanently removes `~/models` directory and all `.gguf` weight files.
  - **Erases vector database**: Permanently deletes `~/knowledge_db.json` and associated vector caches.
- **PRESERVED**:
  - Knowledge notes directory (`agent_knowledge` user notes and documents remain untouched).

```bash
# Execute Level 2 Cleanup (requires typing 'CONFIRM')
wola.purge --medium

# Non-interactive / unattended
wola.purge --medium -y
```

---

#### Level 3: Hard / Nuclear (Complete Self-Destruction)
- **Flag**: `3`, `--hard`, `--nuclear`, `--level 3`, `--level=3`
- **Actions Performed**:
  - Irreversibly erases **EVERY** file, folder, and database created by or related to Wolarc:
    1. Stops all active processes and releases wake-locks.
    2. Deletes all binaries, symlinks, configurations, and runtime state.
    3. Deletes all log files and background server scripts.
    4. Permanently erases `~/models` and all GGUF weights.
    5. Permanently erases `~/knowledge_db.json` vector database.
    6. **Permanently erases the knowledge notes directory** (`$KNOWLEDGE_DIR` / `agent_knowledge`).
    7. **Permanently erases the cloned repository source directory**, leaving no residual trace of Wolarc on the system.

```bash
# Execute Level 3 Self-Destruction (requires typing 'SELF-DESTRUCT')
wola.purge --hard

# Non-interactive / unattended execution
wola.purge --hard -y
```

---

### Command Options Reference

| Flag / Option | Description |
| :--- | :--- |
| `1`, `--soft`, `--level 1` | Run Level 1 Soft cleanup (preserve models, vectors, knowledge) |
| `2`, `--medium`, `--level 2` | Run Level 2 Medium cleanup (erase system files and models, preserve knowledge) |
| `3`, `--hard`, `--nuclear` | Run Level 3 Complete Self-Destruction (erase everything permanently) |
| `-y`, `--yes`, `--force` | Skip interactive text confirmation prompts |
| `--purge-repo` | Explicitly delete repository folder (default in Level 3) |
| `-h`, `--help` | Show command usage and description of the 3 hardness levels |

If invoked without arguments (`wola.purge`), an interactive selector menu is displayed with descriptions of each level before any files are modified.

---

## Planned Commands (Roadmap Milestones) 🗺️

As detailed in [**ROADMAP.md**](ROADMAP.md), the upcoming client-host connection fabric and native Flutter companion app will introduce the following strict 1-to-1 command pairs:

| Planned Feature | Casual Name | Technical Name | Description |
| :--- | :--- | :--- | :--- |
| **Client Pairing** | `wola.pair` | `wola.pair` | Generate cryptographic QR code / PIN for pairing Flutter and remote client apps |
| **Connected Peers** | `wola.clients` | `wola.peers` | Inspect active authorized client devices, token scopes, and connection transports |
| **Connection Diagnostics** | `wola.connection` | `wola.conn` | Benchmark and monitor multi-transport latency (mDNS, LAN, Tailscale, Cloudflare) |


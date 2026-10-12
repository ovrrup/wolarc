# Wolarc 🐺⚡

Turn your Android device (Termux) or Linux machine into a private, autonomous, local RAG AI node with persistent remote access.

`wolarc` orchestrates local GGUF models using `llama.cpp`, an automated background vector indexer for your documents, an OpenAI-compatible API gateway, real-time telemetry, and secure tunneling for remote client apps.

---

> 📖 **Guides & Documentation**:
> - [**System Architecture & Reference Manual (`DOCUMENTATION.md`)**](DOCUMENTATION.md) — Comprehensive architecture, RAG vector pipeline, offline hotspot/LAN connectivity, daemon lifecycles, and 1-to-1 command specifications.
> - [**Installation & Complete User Guide (`GETTING_STARTED.md`)**](GETTING_STARTED.md) — Step-by-step installation on Termux & Linux, model downloads, RAG setup, remote client app pairing (Chatbox, Obsidian, NextChat), and troubleshooting.
> - [**Complete Command Reference (`COMMANDS.md`)**](COMMANDS.md) — Exhaustive CLI reference, Technical/Casual command pairs, TUI shortcuts, configuration variables, and discrete system operations.
> - [**Future Milestones & System Roadmap (`ROADMAP.md`)**](ROADMAP.md) — Strategic roadmap detailing resilient client-host connection methods and the official native Flutter cross-platform companion app.

---

## Features

- **Universal Multi-Format RAG & Dialogue Memory Engine**: Background daemon (`indexer.py`) automatically vectorizes files in both your knowledge folder and conversation history folder (`CHATS_DIR`) — documents (`.pdf`, `.docx`, `.epub`, `.md`, `.txt`), chat dialogue turns, audio tracks & transcripts (`.mp3`, `.wav`, `.m4a`, `.ogg`, `.flac`), video metadata & dialogue (`.mp4`, `.mkv`, `.mov`), photos & diagrams with EXIF & OCR (`.jpg`, `.png`, `.svg`), office files (`.pptx`, `.xlsx`), source code, and data tables.
- **Autonomous Reasoning with Long-Term Memory Blending**: The model prioritizes its own foundational intelligence and creative reasoning instead of robotic verbatim quoting. Verified personal knowledge and past conversational dialogues are seamlessly retrieved as background memory to deliver tailored answers without awkward "according to the notes" disclaimers.
- **Configurable Thinking Mechanism (`wola think [on|off]`)**: Dynamically toggle model reasoning `<think>` tokens on or off to conserve battery, CPU compute, and response latency across Web UI, Go TUI, CLI (`wola.chat`), and API.
- **Industry-Standard Go & Bubble Tea TUI (`wola-tui`)**: Fully redesigned in Go using Charm Bubble Tea, Lip Gloss, and Bubbles with animations and micro-transitions:
  - `⚡ Dashboard`: Real-time hardware telemetry, live Unicode velocity sparklines (` ▂▃▄▅▆▇█`), multi-color gradient bars, animated daemon spinners, and glowing radar pulse indicators.
  - `💬 Chat`: Multi-turn conversational chat with live SSE token streaming, token velocity badges (`[⚡ 19.4 tok/s • 342 tok]`), collapsible reasoning blocks (`Ctrl+T`), and dynamic RAG injection.
  - `📦 Models`: Local GGUF model manager (`~/models`), active role assignments, and external storage importer (Move/Copy).
  - `📚 Knowledge`: Vector database inspector, document summaries, chunk counts, and live semantic search retrieval tester with gradient match bars.
  - `📋 Logs`: Live multi-daemon log tailer with auto-scroll and stream selector (`chat`, `embed`, `agent`, `tunnel`, `indexer`, `thinking`).
  - `⚡ Command Palette (`Ctrl+K` / `:`)`: Instant fuzzy action launcher for service management, view jumping, and URL copying.
  - `❓ Help Overlay Modal (`?` / `F1`)`: Floating cheat sheet with hotkeys, navigation guide, and tips.
  - `🍞 Micro-Toast System`: Non-intrusive transient notifications with automatic dismissal for system events and clipboard sync.
- **Dual Engine Architecture**: Concurrently runs an embedding engine and a chat LLM engine via `llama-server`.
- **OpenAI-Compatible Gateway**: Exposes `/v1/chat/completions` and `/v1/models` (port 5000), supporting real-time SSE streaming.
- **Standalone Zero-Dependency Web Chat UI**: Built-in browser interface served directly from the local gateway (`http://<ip>:5000/`) with zero external CDNs or remote font dependencies. Includes markdown rendering, copy buttons, live SSE token streaming, thinking on/off toggle button, and collapsible cognitive reasoning process boxes (`<think>`).
- **Air-Gapped Offline Gateway & Hotspot Pairing (`wola.local` / `wola.hotspot`)**: Connect completely offline over local Wi-Fi, Android portable hotspot (`192.168.43.1`), or USB tethering (`192.168.42.129`) with auto-generated terminal QR codes and `/v1/local_info` discovery.
- **Tailscale WireGuard Mesh VPN (`wola.tailscale` / `wola.ts`)**: Direct peer-to-peer encrypted mesh connectivity from any device worldwide without public relays or port forwarding. Automatically detects the Android OS Tailscale VPN app or Linux CLI, supports MagicDNS hostnames, Serve & Funnel, and injects universal CORS headers.
- **Autonomous Tunnel Watchdog Daemon (`bin/wola-tunnel-daemon`)**: Self-healing background supervisor monitoring tunnel health every 20s. Automatically cascades across Cloudflare Quick Tunnels (`trycloudflare.com`), `localhost.run`, `serveo.net`, and `pinggy.io` upon connection degradation.
- **Multi-Route Reachability & Latency Benchmark (`wola.link test`)**: Simultaneously tests and benchmarks latency across Localhost, Wi-Fi LAN, Mobile Hotspot, Tailscale mesh, and Public tunnels with outbound ISP reachability sweeps.
- **Dual TUI / CLI Operational Toggle (`wola tui [on|off]`)**: Seamlessly switch between the full-screen Bubble Tea TUI and the classic script-based terminal suite.
- **Real-Time Thinking Viewer & Switch (`wola.th` / `wola think`)**: Inspect model reasoning (`<think>` tokens) and RAG retrieval scores live as inference streams, or turn off thinking for instant response velocity.
- **In-Terminal Model Store & Importer (`wola.gt` / `wola.gh`)**: Search and download tested GGUF models directly from Hugging Face or import your own local `.gguf` files with automatic Move/Copy management and role assignment.
- **Zero-Config Remote Tunneling**: Automatically establishes an encrypted tunnel via Cloudflare, `localhost.run`, `serveo.net`, or `pinggy.io` for remote access on phone or PC.
- **Smart Setup & Conflict Engine**: `install.sh` detects existing processes/configurations and offers Clean Slate (`--clean`), Adjust & Merge BETA (`--adjust`), or Safe Upgrade (`--upgrade`).
- **Process & Power Management**: Integrated Termux wake-lock support and safe process lifecycle controls (`wola.r`, `wola.k`, `wola.rs`).
- **CI/CD Automated Cross-Compilation**: GitHub Actions workflow builds multi-architecture binaries (`linux/arm64`, `linux/amd64`, `darwin/arm64`, `windows/amd64`) on every push.

---

## Architecture Overview

```
                      ┌──────────────────────────────────────────────┐
                      │             Remote Client Apps               │
                      │  (Chatbox, Obsidian, NextChat, OpenWebUI)    │
                      └──────────────────────┬───────────────────────┘
                                             │
                       ┌─────────────────────┼─────────────────────┐
                       │                     │                     │
                Tailscale Mesh VPN     Cloudflare / SSH      Local Network
                 (Encrypted P2P)       (Watchdog Tunnel)    (Hotspot / LAN)
                       │                     │                     │
                       └─────────────────────┼─────────────────────┘
                                             ▼
                               ┌───────────────────────────┐
                               │   Agent API Gateway       │
                               │   (agent.py / Port 5000)  │
                               └──────┬─────────────┬──────┘
                                      │             │
                Query Vector / Cosine │             │ Augmented Prompt
                                      ▼             ▼
  ┌────────────────────────┐   ┌──────────────┐   ┌──────────────┐
  │ Local Notes & Docs     │──>│ Embed Server │   │  Chat LLM    │
  │ (indexer.py background)│   │ (Port 8081)  │   │ (Port 8080)  │
  │ └─ SHA-256 Deduplication   │ (Port 8081)  │   │              │
  └────────────────────────┘   └──────────────┘   └──────────────┘
```

---

## Quick Start (Zero Code Know-How Required) 🚀

### ⚡ 1-Command Installation

Just open Termux (Android) or your Linux terminal, copy and paste this command, and press Enter:

```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

> **What the installer handles automatically for you:**
> - **Zero-Setup AI Engine**: Automatically installs and configures `llama-server` from default (`pkg install llama-cpp` on Android/Termux, or official prebuilt binary on Linux) — **zero manual compilation required**.
> - **Automated Dependencies**: Installs Python, Git, Curl, Go, and all required packages.
> - **Starter Model Acquisition**: Automatically downloads and configures lightweight, 100% free and public starter models (chat & vector embedding).
> - **Modern Interface**: Compiles the Go Bubble Tea TUI (`wola`) and command shortcuts (`wola.r`, `wola.chat`, `wola.url`).

---

### Manual Clone (Alternative)

```bash
git clone https://github.com/ovrrup/wolarc.git
cd wolarc
bash install.sh
```

The installer features interactive permission prompts and custom modes:
- **Llama Inference Engine Prompt (`--llama` / `--no-llama`)**: Asks whether to install `llama-server` and engine files.
- **Experimental Go TUI Prompt (`--tui` / `--no-tui`)**: Asks whether to install the experimental Go Bubble Tea TUI or continue with the lightweight classic CLI suite without it.
- **[1] Clean Slate (`--clean`)**: Stops running services, cleans stale locks/logs, removes previous binaries while preserving downloaded models.
- **[2] Adjust & Merge BETA (`--adjust`)**: Automatically reconciles paths, settings, and databases with existing setups.
- **[3] Safe Upgrade (`--upgrade`)**: Seamlessly updates scripts and binaries without touching user configs.

### 3. Download or Import Models

Use the interactive model store or import local models from custom storage:
```bash
wola.gt          # or wola.gh
# Or import a local file directly:
wola.gt import /path/to/my-model.gguf
# Or via model manager:
wola.m import /path/to/my-model.gguf
```

Choose from lightweight, mobile-tested models such as:
- **Chat**: Gemma 2 2B, Qwen 2.5 1.5B, Llama 3.2 1B, Phi-4 Mini
- **Embedding**: BGE-Small English v1.5, Nomic Embed v1.5, Embedding-Gemma 270M

### 4. Launch or Start the Node

Launch the modern Go TUI:
```bash
wola
```

Or boot background daemons directly:
```bash
wola.r
```

This boots:
1. Embedding Engine (`127.0.0.1:8081`)
2. Chat LLM Engine (`127.0.0.1:8080`)
3. Document Indexer Daemon
4. Agent API Gateway (`127.0.0.1:5000`)
5. Public Remote Tunnel

---

## Command Reference (Strict 1-to-1 Technical & Casual Pairs)

Wolarc enforces a strict 1-to-1 command pairing convention: every feature has **one Casual Name** (friendly, readable) and **one Technical Name** (terse, developer-friendly). Redundant aliases have been eliminated.

| Casual Name (Friendly) | Technical Name (Terse) | Dispatcher Commands | Action Description |
|---|---|---|---|
| `wola` | `wola-tui` | `wola` | **Unified Go & Bubble Tea TUI** (`dashboard`, `chat`, `models`, `knowledge`, `logs`) |
| `wola.start` | `wola.r` | `wola start` / `wola r` | **Start** all services (acquires CPU wake-lock) |
| `wola.stop` | `wola.k` | `wola stop` / `wola k` | **Stop** all services cleanly (releases CPU wake-lock) |
| `wola.restart` | `wola.rs` | `wola restart` / `wola rs` | **Restart** all services gracefully |
| `wola.check` | `wola.status` | `wola check` / `wola status` | View active process PIDs, ports, and switch status |
| `wola.dashboard` | `wola.d` | `wola dashboard` / `wola d` | Open live **telemetry dashboard** (or `--once` for instant snapshot) |
| `wola.chat` | `wola.c` | `wola chat` / `wola c` | **Interactive CLI chat** or single query (`wola.chat "query"`) |
| `wola.models` | `wola.m` | `wola models` / `wola m` | List models in `~/models` or switch active Chat/Embed roles |
| `wola.store` | `wola.gt` | `wola store` / `wola gt` | Open **model store** & local GGUF importer |
| `wola.knowledge` | `wola.i` | `wola knowledge` / `wola i` | Manage vector index (`status`, `add <f>`, `search <q>`) |
| `wola.bench` | `wola.b` | `wola bench` / `wola b` | **Hardware benchmark** (tokens/sec) & auto-tuner |
| `wola.log` | `wola.l` | `wola log` / `wola l` | Unified live daemon **log follower** (`chat`, `embed`, `tunnel`) |
| `wola.think` | `wola.th` | `wola think` / `wola th` | Real-time **thinking stream viewer** or toggle (`on`/`off`/`status`) |
| `wola.link` | `wola.url` | `wola link` / `wola url` | **Remote tunnel URL**, latency test, ASCII QR code & clipboard |
| `wola.hotspot` | `wola.local` | `wola hotspot` / `wola local`| **Offline gateway & Web Chat UI**, LAN/Hotspot links, QR pairing |
| `wola.tailscale` | `wola.ts` | `wola tailscale` / `wola ts` | **Tailscale WireGuard Mesh VPN**, peer access, MagicDNS, Serve & Funnel |
| `wola.config` | `wola.t` | `wola config` / `wola t` | View and edit configuration parameters (`wola.t KEY VALUE`) |
| `wola.update` | `wola.u` | `wola update` / `wola u` | **Check for updates** & upgrade to latest release |
| `wola.uninstall` | `wola.purge` | `wola uninstall` / `wola purge`| Multi-level system cleanup & self-destruct uninstaller |
| `wola.help` | `wola.h` | `wola help` / `wola h` | Interactive **command center** and quick launcher wizard |
| `wola.ui` | `wola.tui` | `wola ui` / `wola tui` | **Toggle default UI mode** between Go TUI and classic CLI |

---

## Tuning Knobs (`wola.t`)

Dynamic runtime knobs (instant effect, no restart needed):
```bash
wola.t TEMPERATURE 0.6              # Adjust sampling temperature (0.6 optimal for natural intelligence)
wola.t SIMILARITY_THRESHOLD 0.50    # High-confidence vector matching threshold
wola.t TOP_K 2                      # Number of context chunks injected
wola.t SYSTEM_PROMPT "..."          # Custom system persona
wola.t ENABLE_THINKING true         # Toggle reasoning output (true: reasoning tokens, false: direct fast)
wola.t ENABLE_TUI true              # Default UI mode (true: Go TUI, false: classic CLI)
```

Engine knobs (applied on next `wola.r` startup):
```bash
wola.t CHAT_THREADS 4               # CPU threads for chat inference
wola.t CHAT_CTX 4096                # Context window size in tokens
wola.t FLASH_ATTENTION auto         # Flash Attention (auto / on / off)
wola.t MLOCK false                  # Pin model in RAM (true / false)
wola.t CHAT_PARALLEL 2              # Parallel slots (prevents remote / local tool lockouts)
wola.t TUNNEL_PROVIDER auto         # auto (fallback), localhost.run, serveo, or pinggy
```

---

## Multi-Provider Tunneling & Remote Access

Wola includes a self-healing multi-provider SSH tunnel engine that automatically falls back across providers if one is unreachable or rate-limited:
1. `localhost.run`
2. `serveo.net`
3. `pinggy.io`

- **Instant URL Delivery**: The active URL is stored in `~/.wola_url` and accessible anytime via `wola.url`.
- **In-Terminal QR Code**: Run `wola.url` to display an ASCII QR code to scan with your phone camera.
- **Clipboard Sync**: Automatically copies to your Android clipboard on Termux.
- **Provider Switching**: Run `wola.url switch serveo` or `wola.url switch pinggy` to switch tunnel routes on the fly.

---

## Connecting Client Apps

Connect any OpenAI-compatible client (Chatbox, Obsidian, NextChat, OpenWebUI):

- **Base URL**: `https://<your-subdomain>.lhr.life/v1` (displayed via `wola.url`)
- **API Key**: `sk-test` (or any non-empty string)
- **Model**: Auto-detected or enter active model name

---

## Future Milestones & System Roadmap 🗺️

Wolarc is expanding into an interconnected local AI ecosystem. See [**ROADMAP.md**](ROADMAP.md) for full architecture specifications, protocols, and phased delivery plans:

### 1. Reliable & Effective Client-Host Connection Fabric
- **Zero-Config Local Auto-Discovery (mDNS / Zeroconf)**: Broadcasts `_wolarc._tcp.local` so clients discover Wolarc nodes instantly on Wi-Fi and hotspots without manual IP entry.
- **Resilient Multi-Transport Fallback**: Automatic route switching across LAN Wi-Fi (<2ms), Tailscale WireGuard direct mesh P2P, persistent Cloudflare Named Tunnels, and WebRTC DataChannels (immune to CGNAT).
- **Resumable Token Streaming**: In-memory event sequence buffer (`Last-Event-ID`) allowing clients to survive network transitions (e.g. Wi-Fi dropping to 5G) without losing tokens or aborting generation.
- **Cryptographic Out-of-Band Pairing**: Dynamic QR code / PIN pairing with Ed25519 key exchange and scoped tokens (`chat`, `knowledge`, `admin`) via `wola.pair` / `wola.peers`.

### 2. Official Native Flutter Cross-Platform App (`wola-flutter`)
- **Cross-Platform Availability**: High-performance native app for Android (APK), iOS, Linux, Windows, and macOS.
- **Next-Gen Chat & Reasoning UI**: SSE token streaming with expandable, styled `<think>` cognitive reasoning blocks, duration counters, token velocity speedometer (`⚡ tok/s`), and full markdown syntax highlighting.
- **One-Tap Pairing Hub**: Built-in camera QR scanner, LAN node discovery, multi-node manager (switch between phone, desktop, and server nodes), and live latency indicator.
- **Universal Knowledge Uploader & Inspector**: Pick documents and media directly from mobile storage to push to the node's vector indexer; browse document chunks and inspect retrieval match percentages.
- **Hardware Telemetry & Node Remote Controls**: Real-time gauges for CPU, RAM, battery, thermal temperature, daemon lifecycle buttons (Start, Stop, Restart), remote model downloader, and hyperparameter tuning sliders.

---

## License

MIT License. See [LICENSE](LICENSE) for details.


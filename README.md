# Wolarc 🐺⚡

Turn your Android device (Termux) or Linux machine into a private, autonomous, local RAG AI node with persistent remote access.

`wolarc` orchestrates local GGUF models using `llama.cpp`, an automated background vector indexer for your documents, an OpenAI-compatible API gateway, real-time telemetry, and secure tunneling for remote client apps.

---

> 📖 **Guides & Documentation**:
> - [**Installation & Complete User Guide (`GETTING_STARTED.md`)**](GETTING_STARTED.md) — Step-by-step installation on Termux & Linux, model downloads, RAG setup, remote client app pairing (Chatbox, Obsidian, NextChat), and troubleshooting.
> - [**Complete Command Reference (`COMMANDS.md`)**](COMMANDS.md) — Exhaustive CLI reference, TUI shortcuts, configuration variables, and discrete system operations.

---

## Features

- **Autonomous RAG Engine**: Background daemon (`indexer.py`) automatically indexes notes (`.md`, `.txt`, `.pdf`) into local vector embeddings using cosine similarity.
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
- **Real-Time Thinking Viewer (`wola.th`)**: Inspect model reasoning (`<think>` tokens) and RAG retrieval scores live as inference streams.
- **In-Terminal Model Store & Importer (`wola.gt` / `wola.gh`)**: Search and download tested GGUF models directly from Hugging Face or import your own local `.gguf` files with automatic Move/Copy management and role assignment.
- **Zero-Config Remote Tunneling**: Automatically establishes an encrypted tunnel via `localhost.run`, `serveo.net`, or `pinggy.io` for remote access on phone or PC.
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
                                             │ HTTPS (localhost.run)
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

## Quick Start

### 1. Requirements

- **Android** with [Termux](https://termux.dev) or any **Linux** environment.
- Python 3.8+
- `llama-server` (built from `llama.cpp`)
- `git`, `curl`, `openssh`

Install dependencies in Termux:
```bash
pkg update && pkg install python git curl openssh poppler
pip install -r requirements.txt
```

### 2. Installation (One-Command Setup)

Install Wolarc and compile the Go TUI automatically with a single command:
```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

Or clone manually:
```bash
git clone https://github.com/ovrrup/wolarc.git
cd wolarc
bash install.sh
```

The installer detects existing files and processes and lets you choose between:
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

## Command Reference

| Command | Action |
|---|---|
| `wola` | **Unified Go & Bubble Tea TUI** (`dashboard`, `chat`, `models`, `knowledge`, `logs`) |
| `wola.r` | **Start** all services (acquires wake-lock) |
| `wola.k` | **Stop** all services cleanly (releases wake-lock) |
| `wola.rs` | **Restart** all services |
| `wola-tui` | Direct binary invocation for **Go & Bubble Tea TUI** |
| `wolarc status` | View active process PIDs and listening ports |
| `wola.d` | Open the live **telemetry dashboard** (auto-delegates to `wola-tui` when present) |
| `wola.url` | **Remote tunnel URL**, latency test, ASCII QR code & clipboard |
| `wola.chat` | **Interactive in-terminal chat** with live token stream & RAG notes |
| `wola.bench` | **Hardware benchmark** (tokens/sec) & optimal thread auto-tuner |
| `wola.log` | Unified live daemon **log follower** (`chat`, `embed`, `tunnel`) |
| `wola.th` | Open the real-time **thinking stream viewer** |
| `wola.gt` / `wola.gh` | Open the **model store** & **local GGUF importer** (Hugging Face / custom storage) |
| `wola.m` | List models in `~/models`, switch models, or import local GGUF (`wola.m import <path>`) |
| `wola.t` | View and edit configuration parameters (`wola.t KEY VALUE`) |
| `wola.i` | Manage vector index (`wola.i status`, `wola.i add <f>`, `wola.i search <q>`) |
| `wola.u` | **Check for updates** & upgrade to latest release (`wola.u`) |
| `wola-h` | Interactive **command center** and quick launcher |

---

## Tuning Knobs (`wola.t`)

Dynamic runtime knobs (instant effect, no restart needed):
```bash
wola.t TEMPERATURE 0.2              # Adjust sampling temperature
wola.t SIMILARITY_THRESHOLD 0.70    # Strictness of vector matching
wola.t TOP_K 3                      # Number of context chunks injected
wola.t SYSTEM_PROMPT "Be concise."  # Assistant persona
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

## License

MIT License. See [LICENSE](LICENSE) for details.

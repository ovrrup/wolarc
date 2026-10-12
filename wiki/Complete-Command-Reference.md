# Complete Command Reference ⌨️

Wolarc enforces a strict **1-to-1 Technical vs. Casual Pairing Convention**:
- **One Casual Name**: Human-readable, intuitive, and easy to remember (e.g. `wola.start`, `wola.knowledge`).
- **One Technical Name**: Terse, fast to type in shells and automation scripts (e.g. `wola.r`, `wola.i`).
- **Unified Dispatcher**: Both conventions can also be invoked via the root `wola` command (e.g. `wola start` or `wola r`). Redundant legacy aliases have been completely eliminated.

---

## 📋 Complete Command Pairing Matrix

| Casual Name (Friendly) | Technical Name (Terse) | Dispatcher Equivalent | Action Description |
|:---|:---|:---|:---|
| `wola` | `wola-tui` | `wola` | **Unified Go & Bubble Tea TUI** (`dashboard`, `chat`, `models`, `knowledge`, `logs`) |
| `wola.ui` | `wola.tui` | `wola ui` / `wola tui` | **Toggle default UI mode** between Go TUI and classic CLI (`on` / `off` / `status`) |
| `wola.start` | `wola.r` | `wola start` / `wola r` | **Start** all services (acquires CPU wake-lock) |
| `wola.stop` | `wola.k` | `wola stop` / `wola k` | **Stop** all services cleanly (releases CPU wake-lock) |
| `wola.restart` | `wola.rs` | `wola restart` / `wola rs` | **Restart** all services gracefully |
| `wola.check` | `wola.status` | `wola check` / `wola status` | View active process PIDs, ports, and switch status |
| `wola.dashboard` | `wola.d` | `wola dashboard` / `wola d` | Open live **telemetry dashboard** (or `--once` for instant snapshot) |
| `wola.chat` | `wola.c` | `wola chat` / `wola c` | **Interactive CLI chat** or single query (`wola.chat "query"`) |
| `wola.models` | `wola.m` | `wola models` / `wola m` | List models in `~/models`, switch active roles, or delete (`rm`) |
| `wola.store` | `wola.gt` | `wola store` / `wola gt` | Open **model store** & local GGUF importer |
| `wola.knowledge` | `wola.i` | `wola knowledge` / `wola i` | Manage vector index (`status`, `add`, `search`, `rm`, `rm-chat`, `clear-chats`) |
| `wola.bench` | `wola.b` | `wola bench` / `wola b` | **Hardware benchmark** (tokens/sec) & auto-tuner |
| `wola.log` | `wola.l` | `wola log` / `wola l` | Unified live daemon **log follower** (`chat`, `embed`, `tunnel`, `clear`) |
| `wola.think` | `wola.th` | `wola think` / `wola th` | Real-time **thinking stream viewer** or toggle (`on`/`off`/`status`) |
| `wola.link` | `wola.url` | `wola link` / `wola url` | **Remote tunnel URL**, latency test, ASCII QR code & clipboard |
| `wola.hotspot` | `wola.local` | `wola hotspot` / `wola local`| **Offline gateway & Web Chat UI**, LAN/Hotspot links, QR pairing |
| `wola.tailscale` | `wola.ts` | `wola tailscale` / `wola ts` | **Tailscale WireGuard Mesh VPN**, peer access, MagicDNS, Serve & Funnel |
| `wola.config` | `wola.t` | `wola config` / `wola t` | View and edit configuration parameters (`wola.t KEY VALUE`) |
| `wola.update` | `wola.u` | `wola update` / `wola u` | **Check for updates** & upgrade to latest release |
| `wola.uninstall` | `wola.purge` | `wola uninstall` / `wola purge`| Multi-level system cleanup & self-destruct uninstaller |
| `wola.help` | `wola.h` | `wola help` / `wola h` | Interactive **command center** and quick launcher wizard |

---

## 🔍 Detailed Command Specifications

### 1. Process Lifecycle
```bash
# Start all background daemons
wola.start
wola.r

# Stop all background daemons and release wake-lock
wola.stop
wola.k

# Graceful restart cycle
wola.restart
wola.rs

# Check running daemon status and listening ports
wola.check
wola.status
```

### 2. Conversational Chat
```bash
# Launch interactive terminal chat session
wola.chat
wola.c

# Execute a one-shot query from command line
wola.chat "What are the primary differences between BGE and Nomic embeddings?"
wola.c "Explain how Flash Attention saves memory bandwidth."
```

### 3. Model Management & Store
```bash
# List all downloaded models and active role bindings
wola.models
wola.m

# Interactively delete a model from ~/models (clears binding if active)
wola.models rm
wola.m rm my-old-model.gguf

# Launch interactive Hugging Face model store
wola.store
wola.gt

# Import a local GGUF file from storage (interactive Move or Copy)
wola.store import /sdcard/Download/qwen2.5-1.5b.gguf
wola.m import /sdcard/Download/qwen2.5-1.5b.gguf
```

### 4. Vector Knowledge Base (RAG)
```bash
# View vector database status, chunk counts, and storage size
wola.knowledge status
wola.i status

# Manually trigger indexing for a file or folder
wola.knowledge add ~/notes/meeting.pdf
wola.i add ~/Documents/wola/knowledge/

# Perform semantic similarity search directly in terminal
wola.knowledge search "quarterly revenue projections"
wola.i search "database concurrency locks"

# Delete a document and atomically prune its vectors from knowledge_db.json
wola.knowledge rm meeting.pdf
wola.i rm meeting.pdf

# Remove a specific chat transcript and prune its dialogue vectors
wola.knowledge rm-chat session_20261012_153022
wola.i rm-chat session_20261012_153022

# Clear all historical chat logs and conversation vectors
wola.knowledge clear-chats
wola.i clear-chats
```

### 5. Diagnostics, Logging & Cognitive Thinking
```bash
# Follow live daemon logs
wola.log chat        # llama-server Chat LLM
wola.log embed       # llama-server Embed Engine
wola.log agent       # Agent API Gateway
wola.log tunnel      # Remote Watchdog Tunnel
wola.log think       # Cognitive reasoning (<think>) stream

# Truncate all log files to 0 bytes and reclaim storage
wola.log clear
wola.log clean

# View or toggle reasoning tokens dynamically
wola.think status    # Check if thinking output is enabled
wola.think on        # Enable <think> tokens
wola.think off       # Disable <think> tokens for maximum speed
wola.th              # Follow live thinking log
```

### 6. Network Connectivity & Gateways
```bash
# View remote tunnel URL and scan in-terminal ASCII QR code
wola.link
wola.url

# Run latency benchmark across all connection routes
wola.link test
wola.url test

# Switch active tunnel provider on the fly
wola.url switch cloudflare
wola.url switch localhost.run
wola.url switch serveo
wola.url switch pinggy

# Display air-gapped offline hotspot & LAN URLs with pairing QR
wola.hotspot
wola.local

# Tailscale WireGuard Mesh management
wola.tailscale status
wola.ts up
wola.ts ip
```

### 7. Configuration Tuning (`wola.config` / `wola.t`)
```bash
# View entire current configuration
wola.config
wola.t

# View a single parameter
wola.t TEMPERATURE

# Update a parameter dynamically
wola.t TEMPERATURE 0.6
wola.t SIMILARITY_THRESHOLD 0.55
wola.t TOP_K 3
wola.t ENABLE_THINKING true
wola.t CHAT_THREADS 4
```

### 8. System Uninstallation & Granular Purging
```bash
# Interactive uninstaller menu (choose Level 1, 2, or 3)
wola.uninstall
wola.purge

# Targeted granular cleanup flags:
wola.purge --logs      # Purge all log files in ~/.wola/logs/
wola.purge --cache     # Clean __pycache__, locks, and temp files
wola.purge --chats     # Purge chat transcripts and conversation vectors
wola.purge --db        # Purge vector database (re-indexes from scratch)
wola.purge --models    # Purge all models in ~/models/

# Non-interactive executions:
wola.purge --soft -y   # Level 1: Clean system, preserve models & knowledge
wola.purge --hard -y   # Level 3: Full nuclear self-destruction
```

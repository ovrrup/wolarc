# Wolarc Node Suite — Complete Command & Architecture Reference

> **Wolarc** is an autonomous, high-performance, local RAG (Retrieval-Augmented Generation) AI node designed for Termux (Android) and Linux. It exposes an OpenAI-compatible API, indexes local knowledge notes in real time, and establishes resilient public tunnels for remote access.

---

## Command Quick Index

| Command | Aliases | Description |
| :--- | :--- | :--- |
| [`wola.r`](#1-system-lifecycle--process-management) | `wolarc r`, `wolarc start` | Start all AI node services (chat, embed, API, tunnel, wake-lock) |
| [`wola.k`](#1-system-lifecycle--process-management) | `wolarc k`, `wolarc stop` | Stop all AI node services cleanly and release wake-locks |
| [`wola.rs`](#1-system-lifecycle--process-management) | `wolarc rs`, `wolarc restart` | Graceful restart cycle (shutdown -> cooldown -> boot) |
| [`wolarc status`](#1-system-lifecycle--process-management) | — | Display active process PIDs, listening ports, and switch status |
| [`wola.d`](#2-telemetry-dashboard--modern-tui) | `wola.tui`, `wola-tui`, `wolarc d` | Launch the high-performance Go Bubble Tea telemetry dashboard |
| [`wola.url`](#3-remote-gateway--connectivity) | `wola.link`, `wolarc url` | Show public tunnel URL, QR code, ping latency & copy endpoint |
| [`wola.chat`](#4-interactive-terminal-chat-client) | `wola.c`, `wolarc c` | Interactive CLI chat client with live token streaming and RAG citations |
| [`wola.th`](#5-deep-thinking-stream-viewer) | `wolarc th` | Real-time streaming monitor for model reasoning chains (`<think>`) |
| [`wola.gt`](#6-model-store--local-gguf-importer) | `wola.gh`, `wolarc gt` | Model Store & local GGUF file importer (HuggingFace / storage) |
| [`wola.m`](#6-model-store--local-gguf-importer) | `wolarc m` | List models in `~/models` or import `.gguf` files via CLI |
| [`wola.i`](#7-vector-knowledge-base--rag-indexer) | `wolarc i` | Vector indexer manager (`status`, `add <path>`, `search <query>`) |
| [`wola.bench`](#8-hardware-benchmarking--optimization) | `wola.b`, `wolarc b` | Benchmark tokens/sec across CPU threads and auto-tune config |
| [`wola.t`](#9-parameter-tuner--configuration) | `wolarc t` | View and edit runtime inference hyperparameters in `~/.wolarc` |
| [`wola.log`](#10-live-log-monitoring--diagnostics) | `wola.logs`, `wolarc log` | Live multiplexed log viewer for chat, embed, agent, and tunnels |
| [`wola.u`](#11-software-updates--maintenance) | `wola.upgrade`, `wolarc u` | Check for updates, pull releases, and update binaries |
| [`wola-h`](#12-interactive-command-center-wizard) | `wola.h` | Interactive menu wizard with status overview and quick launchers |
| [`wola.purge`](#13-system-file-deletion--self-destruct-engine) | `wola.destroy`, `wola.selfdestruct`, `wola.nuke` | Multi-level file deletion, cleanup, and complete self-destruction |

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

---

## 3. Remote Gateway & Connectivity

### `wola.url`
Manages public network access and remote device pairing.

```bash
# Display public URL, ping latency, and QR code
wola.url

# Reconnect or switch tunnel provider
wola.url reconnect
```

#### Features:
- **Resilient Multi-Provider Fallback**: Automatically cascades across `localhost.run` -> `serveo.net` -> `pinggy.io`.
- **Latency Ping**: Measures HTTP round-trip latency to verify tunnel health.
- **ASCII QR Code**: Generates an interactive terminal QR code for camera scanning from mobile devices.
- **Clipboard Sync**: Automatically copies `https://<tunnel-url>/v1` to the clipboard using `termux-clipboard-set` or `xclip`.
- **Local IP Fallback**: Displays LAN endpoint (e.g. `http://192.168.1.X:5000/v1`) for local Wi-Fi clients.

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

## 5. Deep Thinking Stream Viewer

### `wola.th`
Dedicated real-time reasoning monitor.

```bash
# Stream reasoning tokens live
wola.th
```

Models supporting reasoning (such as DeepSeek-R1, Qwen-2.5-Coder, etc.) write raw cognitive traces into `~/.wola_thinking.log`. `wola.th` provides a color-highlighted live viewer that strips `<think>` XML tags and displays the model's inner monologue in real time before the final answer is rendered.

---

## 6. Model Store & Local GGUF Importer

### `wola.gt` / `wola.gh`
Interactive Model Hub for downloading and managing GGUF models.

```bash
# Launch interactive Model Hub
wola.gt
# or
wola.gh
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
```

---

## 7. Vector Knowledge Base & RAG Indexer

### `wola.i`
CLI tool to manage and test the vector knowledge database.

```bash
# View database health, total indexed chunks, and file count
wola.i status

# Index a single document or an entire directory immediately
wola.i add ~/storage/shared/Documents/my_notes.md
wola.i add /path/to/folder/

# Test vector retrieval directly with a search query
wola.i search "How do I configure network settings?"
```

#### Automatic Indexing Daemon (`indexer.py`):
Runs in the background watching `KNOWLEDGE_DIR` (default: `~/storage/shared/Documents/agent_knowledge`). When files are created, edited, or removed (`.txt`, `.md`, `.json`, `.pdf`), the indexer automatically chunks documents, generates vector embeddings via port `8081`, and saves them to `~/knowledge_db.json`.

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
```

---

## 11. Software Updates & Maintenance

### `wola.u`
Checks for updates from GitHub and upgrades the local installation.

```bash
# Check and apply updates
wola.u
```

- Compares local Git commit hash (`~/.wola_version`) against remote repository HEAD.
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

### `wola.purge`
The system deletion and cleanup engine.

> **Command Aliases**: `wola.purge`, `wola.destroy`, `wola.selfdestruct`, `wola.nuke`, `wolarc purge`, `wolarc destroy`, `wolarc selfdestruct`, `wolarc nuke`.

This command stops all running processes, removes lock files, and erases files and directories associated with Wolarc. It provides **3 distinct hardness levels** to control how thoroughly data is removed.

```bash
# Syntax
wola.purge [OPTIONS] [LEVEL]
```

### Hardness Levels

```
Level 1 (Soft)       ──> Deletes binaries, configs, locks, and logs.
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
  3. Deletes runtime state and lock files (`~/.wola.lock`, `~/.wola_url`, `~/.wola_endpoint`, `~/.wola_provider`, `~/.wola_metrics.json`, `~/.wola_update_cache.json`, `~/.wola_bench_cache.json`, `~/.wola_version`, `~/.wola_thinking.log`).
  4. Deletes log files (`~/chat.log`, `~/embed.log`, `~/agent.log`, `~/tunnel.log`, `~/indexer.log`, `~/llama_chat.log`, `~/llama_embed.log`).
  5. Removes deployed server scripts in `$HOME` (`~/agent.py`, `~/indexer.py`).
  6. Removes configuration file `~/.wolarc` and its backups.
  7. Deletes all installed executables and symlinks in `$TARGET_BIN` (`wolarc`, `wola.r`, `wola.k`, `wola.rs`, `wola.d`, `wola.url`, `wola.chat`, `wola.bench`, `wola.log`, `wola.th`, `wola.gt`, `wola.gh`, `wola.t`, `wola.m`, `wola.i`, `wola.u`, `wola-h`, `wola-tui`).
- **PRESERVED**:
  - `~/models` (downloaded GGUF models are kept intact)
  - `~/knowledge_db.json` (vector database is kept intact)
  - Knowledge notes directory (`agent_knowledge` documents are kept intact)

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

# Quick Start Guide 🚀

Get up and running with your own private, autonomous AI node on Android (Termux) or Linux in under 5 minutes.

---

## 📋 System Requirements

| Specification | Android (Termux) | Linux (x86_64 / aarch64) |
|:---|:---|:---|
| **OS** | Android 9.0+ (Termux from F-Droid or GitHub) | Ubuntu 20.04+, Debian 11+, Arch, Fedora |
| **Architecture** | `aarch64` (ARM64 64-bit) | `x86_64` or `aarch64` |
| **RAM** | 4GB minimum (6GB+ recommended) | 4GB minimum (8GB+ recommended) |
| **Storage** | 3GB free space for models & indexes | 3GB free space |
| **Permissions** | Storage permission (`termux-setup-storage`) & Wake-Lock | Sudo / root for initial dependencies |

> [!IMPORTANT]
> **Termux Note**: Always install Termux from **F-Droid** or the **Termux GitHub releases**. Do NOT use Google Play Store versions as they are outdated and broken.

---

## ⚡ 1-Command Automated Installation

Open your terminal and paste the automated installer:

```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

### What the installer handles automatically:
1. **Inference Engine (`llama-server`)**: Automatically installs `llama-cpp` via `pkg` on Termux, or downloads official prebuilt binaries on Linux. No manual compilation is required.
2. **System Dependencies**: Installs Python 3, Git, Curl, Go compiler, and required libraries.
3. **Starter Model Setup**: Downloads tested, lightweight public models (Chat LLM and Vector Embedding model).
4. **Environment Initialization**: Compiles the Go Bubble Tea TUI (`wola`) and installs all CLI command pairs into your `$PATH`.
5. **Storage Consolidation**: Sets up the centralized `~/.wola/` directory structure (`lib`, `logs`, `run`, `data`).

---

## 🛠️ Manual Installation (Alternative)

```bash
git clone https://github.com/ovrrup/wolarc.git
cd wolarc
bash install.sh
```

### Installer CLI Flags & Custom Modes

The installer supports interactive and unattended non-interactive flags:

```bash
# Safe upgrade (updates binaries & scripts, preserves models and configs)
bash install.sh --upgrade -y

# Clean slate install (restarts daemons, resets logs/runtime, preserves models)
bash install.sh --clean -y

# Adjust & Merge BETA (reconciles paths and config with existing setup)
bash install.sh --adjust -y

# Headless / lightweight setup without Go TUI compilation
bash install.sh --no-tui -y

# Skip llama engine setup if you already have custom binaries
bash install.sh --no-llama -y
```

---

## 📦 Step 2: Download or Import Models

Wolarc uses standard GGUF quantized models placed in `~/models`.

### Option A: Browse Hugging Face In-Terminal Store
Launch the interactive model store:
```bash
wola.gt          # Casual name: wola.store
```
Select lightweight models optimized for mobile/edge compute:
- **Chat Models**: Gemma 2 2B Q4_K_M, Qwen 2.5 1.5B Q4_K_M, Llama 3.2 1B Q4_K_M, Phi-4 Mini.
- **Embedding Models**: BGE-Small English v1.5 (`bge-small-en-v1.5.Q4_K_M.gguf`), Nomic Embed v1.5.

### Option B: Import Local GGUF Files
If you already have GGUF models on your device:
```bash
wola.gt import /sdcard/Download/my-model.gguf
# Or via model manager:
wola.m import /path/to/my-model.gguf
```
You can choose whether to **Move** (saves space) or **Copy** the file into `~/models`.

---

## 🚀 Step 3: Launching Wolarc

### Mode A: Launch the Unified Go TUI
```bash
wola
```
* **Dashboard Tab (`1`)**: Real-time hardware telemetry and daemon status.
* **Chat Tab (`2`)**: Interactive conversational chat with streaming tokens and collapsible `<think>` reasoning traces.
* **Models Tab (`3`)**: View models in `~/models` and toggle active Chat / Embed roles.
* **Knowledge Tab (`4`)**: Inspect vector count, chunk stats, and test semantic search.
* **Logs Tab (`5`)**: Real-time multi-daemon log tailer.

### Mode B: Boot Daemons in the Background (Classic Mode)
```bash
wola.start       # Technical name: wola.r
```
This starts:
1. `llama-server` Chat Engine (`127.0.0.1:8080`)
2. `llama-server` Embed Engine (`127.0.0.1:8081`)
3. Document Indexer Daemon (`indexer.py`)
4. Agent API Gateway (`127.0.0.1:5000`)
5. Public Remote Watchdog Tunnel

Check service health anytime:
```bash
wola.check       # Technical name: wola.status
```

---

## 💬 Step 4: First Interaction

### 1. In-Terminal Interactive Chat
```bash
wola.chat        # Technical name: wola.c
```
Or run a one-shot query:
```bash
wola.chat "Summarize the key advantages of local RAG architecture."
```

### 2. Standalone Zero-Dependency Web Chat UI
Wolarc serves a responsive Web UI directly from the local gateway:
```
http://127.0.0.1:5000/
```
Open it in Chrome, Firefox, or your mobile browser. It features markdown rendering, copy buttons, live SSE streaming, and collapsible reasoning boxes.

---

## 📚 Step 5: Ingesting Knowledge (RAG)

Add documents to your personal knowledge base:
```bash
# Add a single file
wola.knowledge add ~/my_notes.pdf

# Or drop files directly into the knowledge folder:
cp /sdcard/Documents/project_notes.md ~/Documents/wola/knowledge/
```
The background indexer automatically detects, chunks, and vectorizes files into `~/.wola/data/knowledge_db.json`.

Test semantic retrieval:
```bash
wola.knowledge search "quarterly roadmap goals"
```

---

## 🛑 Stopping Services Cleanly

```bash
wola.stop        # Technical name: wola.k
```
This safely stops all child daemons and releases the Android CPU wake-lock.

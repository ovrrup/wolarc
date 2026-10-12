# Architecture & System Topology 🏗️

A comprehensive breakdown of Wolarc's multi-tier subsystem architecture, process boundaries, dataflow pipelines, and storage layouts.

---

## 🗺️ Architectural Topology Diagram

```
                      ┌──────────────────────────────────────────────┐
                      │              Remote Client Apps              │
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
                               │    Agent API Gateway      │
                               │  (agent.py / Port 5000)   │
                               │  • WSGI Server (Gunicorn) │
                               │  • Air-Gapped Web Chat UI │
                               │  • SSE Streaming Handler  │
                               │  • Thinking Filter/Logger │
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

## 🧩 Core Subsystem Components

### 1. Agent API Gateway (`agent.py` / Port 5000)
- **Primary Role**: Orchestrates client requests, semantic search retrieval, context injection, and response streaming.
- **Key Capabilities**:
  - **OpenAI Compatibility**: Exposes standard `/v1/chat/completions`, `/v1/models`, and `/v1/embeddings` routes.
  - **Standalone Web UI**: Serves an air-gapped, zero-dependency browser chat interface directly at `GET /`.
  - **Discovery Interface**: Exposes `GET /v1/local_info` providing device IP, active ports, model names, and network interfaces for client auto-discovery.
  - **Dynamic Cognitive Filtering**: Parses and streams `<think>...</think>` tokens in real time, persisting them to `~/.wola/logs/thinking.log` while allowing users to dynamically toggle thinking on or off.

### 2. Dual Engine Isolation (`llama-server`)
Instead of overloading a single process, Wolarc isolates chat inference from vector generation:
- **Chat Server (`Port 8080`)**:
  - Runs the primary conversation model (e.g. `gemma-2-2b-it.Q4_K_M.gguf`).
  - Configured with `--parallel 2` (parallel slots), `--ctx-size 4096`, and Flash Attention to prevent memory bandwidth saturation.
- **Embedding Server (`Port 8081`)**:
  - Runs dedicated embedding weights (e.g. `bge-small-en-v1.5.Q4_K_M.gguf`).
  - Runs with `--embedding`, computing high-speed 384-dimensional or 768-dimensional document vectors without evicting the chat model's KV cache.

### 3. Background Vector Indexer (`indexer.py`)
- Continuously scans `~/Documents/wola/knowledge/` and `~/Documents/wola/chats/`.
- Extracts text from multi-format files (PDF, DOCX, EPUB, TXT, MD, audio transcripts, EXIF).
- Deduplicates files via SHA-256 signatures so modified documents are re-indexed cleanly while unchanged files are skipped.
- **Cross-Process Advisory Locking**: Uses `fcntl.flock` (`fcntl.LOCK_SH` and `fcntl.LOCK_EX`) on `~/.wola/data/knowledge_db.json.lock` to ensure zero race conditions between `indexer.py` and `agent.py`.

### 4. Autonomous Tunnel Watchdog Daemon (`bin/wola-tunnel-daemon`)
- Runs as a detached background supervisor monitoring remote connectivity.
- Executes health probes against the public endpoint every 20 seconds.
- Automatically handles network disconnection: verifies WAN reachability first before attempting reconnects to avoid burning battery when offline.
- Features multi-provider fallback cascade:
  1. Cloudflare Quick Tunnels (`trycloudflare.com`)
  2. `localhost.run`
  3. `serveo.net`
  4. `pinggy.io`

---

## 📂 Storage Architecture (`~/.wola/`)

To prevent clutter in user home directories, all system files are strictly consolidated into `~/.wola/`:

```
~/.wola/
├── lib/                     # Deployed daemon & utility scripts
│   ├── agent.py
│   ├── indexer.py
│   └── web_ui.py
├── logs/                    # Centralized rotating log files
│   ├── server.log           # Chat LLM server log
│   ├── embed.log            # Embedding server log
│   ├── agent.log            # Agent API gateway log
│   ├── indexer.log          # Vector indexer daemon log
│   ├── tunnel.log           # Remote tunnel output
│   ├── thinking.log         # Real-time model cognitive reasoning log
│   └── watchdog.log         # Tunnel watchdog supervisor log
├── run/                     # Transient runtime state & lockfiles
│   ├── wola_chat.pid        # Chat server PID
│   ├── wola_embed.pid       # Embed server PID
│   ├── wola_agent.pid       # Gateway PID
│   ├── wola_indexer.pid     # Indexer PID
│   ├── wola_tunnel.pid      # Tunnel PID
│   ├── wola_tunnel_url      # Active public URL
│   ├── wola_metrics.json    # Hardware telemetry snapshot
│   └── wola.lock            # Global daemon lifecycle mutex
└── data/                    # Persistent vector database
    ├── knowledge_db.json    # Vector embeddings & document chunks
    └── knowledge_db.json.lock
```

### User Data Separation
User-managed documents and models are kept strictly separate from internal system files:
- **`~/models/`**: Downloaded and imported `.gguf` weights.
- **`~/Documents/wola/knowledge/`**: Personal documents, PDFs, notes, and research papers.
- **`~/Documents/wola/chats/`**: Exported multi-turn conversation logs and chat turns.
- **`~/.wolarc`**: User configuration file containing runtime tuning variables.
